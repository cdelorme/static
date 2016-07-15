package staticmd

import (
	"bufio"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday"
)

var readfile = ioutil.ReadFile
var create = os.Create
var mkdirall = os.MkdirAll
var parseFiles = template.ParseFiles
var walk = filepath.Walk

type ht interface {
	Execute(io.Writer, interface{}) error
}

type Generator struct {
	Book         bool   `json:"book,omitempty"`
	Input        string `json:"input,omitempty"`
	Output       string `json:"output,omitempty"`
	Relative     bool   `json:"relative,omitempty"`
	TemplateFile string `json:"template,omitempty"`

	Logger logger

	version  string
	pages    []string
	template ht
}

func (self *Generator) ior(path string) string {
	return strings.TrimSuffix(strings.Replace(path, self.Input, self.Output, 1), filepath.Ext(path)) + ".html"
}

func (self *Generator) depth(path string) string {
	if self.Relative {
		if rel, err := filepath.Rel(filepath.Dir(path), self.Output); err == nil {
			return rel + string(os.PathSeparator)
		}
	}
	return ""
}

func (self *Generator) walk(path string, file os.FileInfo, err error) error {
	if file != nil && file.Mode().IsRegular() && file.Size() > 0 && isMarkdown(path) {
		self.pages = append(self.pages, path)
	}
	return err
}

func (self *Generator) multi() error {
	navi := make(map[string][]navigation)
	var err error

	for i, _ := range self.pages {
		out := self.ior(self.pages[i])
		dir := filepath.Dir(self.ior(out))
		nav := navigation{}

		if filepath.Dir(out) != self.Output && strings.ToLower(basename(out)) == "index" {
			nav.Title = basename(dir)
			if self.Relative {
				nav.Link = filepath.Join(strings.TrimPrefix(dir, filepath.Dir(dir)+string(os.PathSeparator)), filepath.Base(out))
			} else {
				nav.Link = strings.TrimPrefix(dir, self.Output) + string(os.PathSeparator)
			}
			dir = filepath.Dir(dir)
		} else {
			nav.Title = basename(out)
			if self.Relative {
				nav.Link = strings.TrimPrefix(out, filepath.Dir(out)+string(os.PathSeparator))
			} else {
				nav.Link = strings.TrimPrefix(out, self.Output)
			}
		}

		if _, ok := navi[dir]; !ok {
			navi[dir] = make([]navigation, 0)
			if ok, _ := exists(dir); !ok {
				if err = mkdirall(dir, 0770); err != nil {
					self.Logger.Error("failed to create path: %s, %s", dir, err)
				}
			}
		}

		navi[dir] = append(navi[dir], nav)
	}

	for _, p := range self.pages {
		var markdown []byte
		if markdown, err = readfile(p); err != nil {
			self.Logger.Error("failed to read file: %s, %s", p, err)
			return err
		}

		out := self.ior(p)
		dir := filepath.Dir(out)
		page := page{
			Name:    basename(p),
			Version: self.version,
			Nav:     navi[self.Output],
			Depth:   self.depth(out),
		}

		if dir != self.Output && strings.ToLower(basename(p)) == "index" {
			toc := "\n## Table of Contents:\n\n"
			for i, _ := range navi[dir] {
				toc = toc + "- [" + navi[dir][i].Title + "](" + navi[dir][i].Link + ")\n"
			}
			self.Logger.Debug("table of contents for %s, %s", out, toc)
			markdown = append([]byte(toc), markdown...)
		}

		page.Content = template.HTML(blackfriday.MarkdownCommon(markdown))

		var f *os.File
		if f, err = create(out); err != nil {
			self.Logger.Error("%s\n", err)
			return err
		}
		defer f.Close()

		fb := bufio.NewWriter(f)
		defer fb.Flush()

		if err = self.template.Execute(fb, page); err != nil {
			self.Logger.Error("%s\n", err)
		}
	}

	return err
}

func (self *Generator) single() error {
	content := make([]byte, 0)
	toc := "\n"
	previous_depth := 0
	var err error

	for _, p := range self.pages {
		shorthand := strings.TrimPrefix(p, self.Input+string(os.PathSeparator))
		depth := strings.Count(shorthand, string(os.PathSeparator))
		if depth > previous_depth {
			toc = toc + strings.Repeat("\t", depth-1) + "- " + basename(filepath.Dir(p)) + "\n"
		}
		anchor := strings.Replace(shorthand, string(os.PathSeparator), "-", -1)
		toc = toc + strings.Repeat("\t", depth) + "- [" + basename(p) + "](#" + anchor + ")\n"

		var markdown []byte
		if markdown, err = readfile(p); err != nil {
			self.Logger.Error("failed to read file: %s (%s)", p, err)
			continue
		}

		markdown = append([]byte("\n<a id='"+anchor+"'/>\n\n"), markdown...)
		markdown = append(markdown, []byte("\n[back to top](#top)\n\n")...)
		content = append(content, markdown...)
		previous_depth = depth
	}

	content = append([]byte(toc), content...)

	if ok, _ := exists(self.Output); !ok {
		if err = mkdirall(self.Output, 0770); err != nil {
			self.Logger.Error("failed to create path: %s (%s)", self.Output, err)
			return err
		}
	}

	page := page{
		Version: self.version,
		Content: template.HTML(blackfriday.MarkdownCommon(content)),
	}
	out := filepath.Join(self.Output, "index.html")

	var f *os.File
	if f, err = create(out); err != nil {
		self.Logger.Error("%s\n", err)
		return err
	}
	defer f.Close()

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	if err = self.template.Execute(fb, page); err != nil {
		self.Logger.Error("%s\n", err)
	}

	return err
}

func (self *Generator) Generate() error {
	var err error
	if self.template, err = parseFiles(self.TemplateFile); err != nil {
		self.Logger.Error("%s\n", err)
		return err
	}

	self.version = version(self.Input)
	self.Input, _ = filepath.Abs(self.Input)
	self.Output, _ = filepath.Abs(self.Output)
	self.Input = filepath.Clean(self.Input)
	self.Output = filepath.Clean(self.Output)

	if err := walk(self.Input, self.walk); err != nil {
		self.Logger.Error("%s\n", err)
		return err
	}
	self.Logger.Debug("generator state: %+v", self)

	if self.Book {
		return self.single()
	}
	return self.multi()
}
