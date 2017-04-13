//go:generate go-bindata -pkg static -o templates.go templates/
package static

import (
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var readall = ioutil.ReadAll
var open = os.Open
var create = os.Create
var mkdirall = os.MkdirAll

type logger interface {
	Info(string, ...interface{})
	Debug(string, ...interface{})
	Error(string, ...interface{})
}

type operation func([]byte) []byte

// This is the compiler that collects the list markdown files, a title, the
// input and output paths, and whether to produce multiple files (web mode) or
// to produce a single file (default, book mode).
//
// All public properties are not thread safe, so concurrent execution may yield
// errors if those properties are being modified or accessed in parallel.
type Markdown struct {
	Title    string `json:"title,omitempty"`
	Input    string `json:"input,omitempty"`
	Output   string `json:"output,omitempty"`
	Web      bool   `json:"web,omitempty"`
	Template string `json:"template,omitempty"`
	Version  string `json:"version,omitempty"`
	L        logger `json:"-"`

	err   error
	files []string
}

// This function helps us handle any errors encountered during processing
// without forcing us to immediately terminate.
//
// It also is responsible for logging every error encountered.
//
// If the error is nil it ignores it, thus the last non-nil error will be
// returned, so that the caller knows at least one failure has occurred.
func (m *Markdown) errors(err error) {
	if err == nil {
		return
	}
	m.L.Error(err.Error())
	m.err = err
}

// If the absolute path minus the file extension already exist then we want to
// know so we can avoid redundant processing, overwriting files, and potential
// race conditions.
func (m *Markdown) matches(file string) bool {
	for i := range m.files {
		if strings.TrimSuffix(file, filepath.Ext(file)) == strings.TrimSuffix(m.files[i], filepath.Ext(m.files[i])) {
			return true
		}
	}
	return false
}

// This checks the extension against a list of supported extensions.
func (m *Markdown) valid(file string) bool {
	for i := range extensions {
		if filepath.Ext(file) == extensions[i] {
			return true
		}
	}
	return false
}

// When walking through files we collect errors but do not return them, so that
// the entire operation is not canceled due to a single failure.
//
// If there is an error, the file is a directory, the file is irregular, the
// file does not have a markdown extension, or the file name minus its
// extention is already in our list, then we skip that file.
//
// Thus the first file matched is the only file processed, which is to deal
// with multiple valid markdown extensions for the same file basename.
//
// Each verified file is added to the list of files, which we will process
// after we finish iterating all files.
func (m *Markdown) walk(file string, f os.FileInfo, e error) error {
	m.errors(e)
	if e != nil || f.IsDir() || !f.Mode().IsRegular() || f.Size() == 0 || !m.valid(file) || m.matches(file) {
		return nil
	}
	m.files = append(m.files, file)
	return nil
}

// A way to abstract the process of getting a template
func (m *Markdown) template() (*template.Template, error) {
	if m.Template != "" {
		return template.ParseFiles(m.Template)
	}
	var assetFile = "templates/book.tmpl"
	if m.Web {
		assetFile = "templates/web.tmpl"
	}
	d, e := Asset(assetFile)
	if e != nil {
		return nil, e
	}
	t := template.New("markdown")
	return t.Parse(string(d))
}

// This operation processes each file independently, which includes passing to
// each its own page structure.
//
// In the future, when buffered markdown parsers exist, this should leverage
// concurrency, but the current implementation is bottlenecked at disk IO.
//
// The template is created first, using the compiled bindata by default, or the
// supplied template file if able.
func (m *Markdown) web(o operation) error {
	t, e := m.template()
	if e != nil {
		return e
	}
	for i := range m.files {
		in, e := open(m.files[i])
		if e != nil {
			m.errors(e)
			continue
		}
		b, e := readall(in)
		m.errors(in.Close())
		if e != nil {
			m.errors(e)
			continue
		}
		d := o(b)
		m.errors(mkdirall(filepath.Dir(filepath.Join(m.Output, strings.TrimSuffix(strings.TrimPrefix(m.files[i], m.Input), filepath.Ext(m.files[i]))+".html")), os.ModePerm))
		out, e := create(filepath.Join(m.Output, strings.TrimSuffix(strings.TrimPrefix(m.files[i], m.Input), filepath.Ext(m.files[i]))+".html"))
		if e != nil {
			m.errors(e)
			continue
		}

		if e := t.Execute(out, struct {
			Title   string
			Name    string
			Content template.HTML
			Version string
		}{
			Content: template.HTML(string(d)),
			Title:   m.Title,
			Name:    strings.TrimSuffix(filepath.Base(m.files[i]), filepath.Ext(m.files[i])),
			Version: m.Version,
		}); e != nil {
			m.errors(e)
		}
		m.errors(out.Close())
	}
	return nil
}

// This operation processes each file sequentially, and keeps the bytes for all
// files in memory so it can write the output to a single file.
//
// In the future, it would be best if each file were loaded into a buffered
// markdown parser and piped to a buffered template so that the system could
// avoid storing all bytes in memory.
//
// Once every file has been loaded into a single byte array, we run it through
// the markdown processor `operation`, and pass that into a template which then
// pushes the output to a single file.
func (m *Markdown) book(o operation) error {
	t, e := m.template()
	if e != nil {
		return e
	}
	var b []byte
	for i := range m.files {
		in, e := open(m.files[i])
		if e != nil {
			m.errors(e)
			continue
		}
		d, e := readall(in)
		m.errors(in.Close())
		if e != nil {
			m.errors(e)
			continue
		}
		b = append(b, d...)
	}
	d := o(b)
	m.errors(mkdirall(filepath.Dir(m.Output), os.ModePerm))
	out, e := create(m.Output)
	if e != nil {
		return e
	}
	defer out.Close()
	return t.Execute(out, struct {
		Title   string
		Content template.HTML
		Version string
	}{
		Content: template.HTML(string(d)),
		Title:   m.Title,
		Version: m.Version,
	})
}

// The primary function, which accepts the operation used to convert markdown
// into html.  Unfortunately there are currently no markdown parsers that
// operate on a stream, but in the future I would like to switch to an
// io.Reader interface.
//
// The operation begins by capturing the input path so that we can translate
// the output path when creating files from the input path, including matching
// directories.
//
// If no title has been supplied it will default to the parent directories
// name, but this might be better placed in package main.
//
// The default output for web is `public/`, otherwise when in book mode the
// default is the title.
//
// We walk the input path, which assembles the list of markdown files and then
// we gather any errors returned.
//
// Finally we process the files according to the desired output mode.
func (m *Markdown) Run(o operation) error {
	var e error
	if m.Input == "" {
		if m.Input, e = os.Getwd(); e != nil {
			m.errors(e)
			return e
		}
	}
	if m.Title == "" {
		m.Title = filepath.Base(filepath.Dir(m.Input))
	}
	if m.Web && m.Output == "" {
		m.Output = filepath.Join(m.Input, "public")
	} else if m.Output == "" {
		m.Output = filepath.Join(m.Input, m.Title+".html")
	}
	m.errors(filepath.Walk(m.Input, m.walk))
	m.L.Debug("Status: %#v", m)
	if m.Web {
		m.errors(m.web(o))
	} else {
		m.errors(m.book(o))
	}
	return m.err
}
