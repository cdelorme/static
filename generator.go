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
	Input        string
	Output       string
	TemplateFile string
	Book         bool
	Relative     bool
	Logger       logger

	version  string
	pages    []string
	template ht
}

// convert markdown input path to html output path
// @note: we cannot reverse since we do not track the extension
// we support `.md`, `.mkd`, and `.markdown`
func (self *Generator) ior(path string) string {
	return strings.TrimSuffix(strings.Replace(path, self.Input, self.Output, 1), filepath.Ext(path)) + ".html"
}

// for relative rendering generate relative depth string, or fallback to empty
func (self *Generator) depth(path string) string {
	if self.Relative {
		if rel, err := filepath.Rel(filepath.Dir(path), self.Output); err == nil {
			return rel + string(os.PathSeparator)
		}
	}
	return ""
}

// walk the directories and build a list of markdown files with content
func (self *Generator) walk(path string, file os.FileInfo, err error) error {
	if file != nil && file.Mode().IsRegular() && file.Size() > 0 && isMarkdown(path) {
		self.pages = append(self.pages, path)
	}
	return err
}

// build output concurrently to many pages
func (self *Generator) multi() (err error) {

	// prepare navigation storage
	navi := make(map[string][]navigation)

	// loop pages to build table of contents
	for i, _ := range self.pages {

		// build output directory
		out := self.ior(self.pages[i])
		dir := filepath.Dir(self.ior(out))

		// create navigation object
		nav := navigation{}

		// sub-index condition changes name, dir, and link
		if filepath.Dir(out) != self.Output && strings.ToLower(basename(out)) == "index" {

			// set name to containing folder
			nav.Title = basename(dir)

			// set relative or absolute link
			if self.Relative {
				nav.Link = filepath.Join(strings.TrimPrefix(dir, filepath.Dir(dir)+string(os.PathSeparator)), filepath.Base(out))
			} else {
				nav.Link = strings.TrimPrefix(dir, self.Output) + string(os.PathSeparator)
			}

			// update dir to dir of dir
			dir = filepath.Dir(dir)
		} else {

			// set name to files name
			nav.Title = basename(out)

			// set relative or absolute link
			if self.Relative {
				nav.Link = strings.TrimPrefix(out, filepath.Dir(out)+string(os.PathSeparator))
			} else {
				nav.Link = strings.TrimPrefix(out, self.Output)
			}
		}

		// build indexes first-match
		if _, ok := navi[dir]; !ok {
			navi[dir] = make([]navigation, 0)

			// create output directory for when we create files
			if ok, _ := exists(dir); !ok {
				if err = mkdirall(dir, 0770); err != nil {
					self.Logger.Error("Failed to create path: %s, %s", dir, err)
				}
			}
		}

		// append to navigational list
		navi[dir] = append(navi[dir], nav)
	}

	// process all pages
	for _, p := range self.pages {

		// attempt to read entire document
		var markdown []byte
		if markdown, err = readfile(p); err != nil {
			self.Logger.Error("failed to read file: %s, %s", p, err)
			return
		}

		// acquire output filepath
		out := self.ior(p)
		dir := filepath.Dir(out)

		// prepare a new page object for our template to render
		page := page{
			Name:    basename(p),
			Version: self.version,
			Nav:     navi[self.Output],
			Depth:   self.depth(out),
		}

		// if this page happens to be a sub-index we can generate the table of contents
		if dir != self.Output && strings.ToLower(basename(p)) == "index" {

			// iterate and build table of contents as basic markdown
			toc := "\n## Table of Contents:\n\n"
			for i, _ := range navi[dir] {
				toc = toc + "- [" + navi[dir][i].Title + "](" + navi[dir][i].Link + ")\n"
			}

			// debug: table of contents output
			self.Logger.Debug("table of contents for %s, %s", out, toc)

			// prepend table of contents
			markdown = append([]byte(toc), markdown...)
		}

		// convert to html, and accept as part of the template
		page.Content = template.HTML(blackfriday.MarkdownCommon(markdown))

		// attempt to open file for output
		var f *os.File
		if f, err = create(out); err != nil {
			return err
		}
		defer f.Close()

		// prepare a writer /w buffer
		fb := bufio.NewWriter(f)
		defer fb.Flush()

		// attempt to use template to write output with page context
		err = self.template.Execute(fb, page)
	}

	return
}

// build output synchronously to a single page
func (self *Generator) single() (err error) {

	// prepare []byte array to store all files markdown
	content := make([]byte, 0)

	// prepare a table-of-contents
	toc := "\n"

	previous_depth := 0

	// iterate and append all files contents
	for _, p := range self.pages {

		// shorthand
		shorthand := strings.TrimPrefix(p, self.Input+string(os.PathSeparator))

		// acquire depth
		depth := strings.Count(shorthand, string(os.PathSeparator))

		// if depth > previous depth then prepend with basename of dir for sub-section-headings
		if depth > previous_depth {
			toc = toc + strings.Repeat("\t", depth-1) + "- " + basename(filepath.Dir(p)) + "\n"
		}

		// prepare anchor text
		anchor := strings.Replace(shorthand, string(os.PathSeparator), "-", -1)

		// create new toc record
		toc = toc + strings.Repeat("\t", depth) + "- [" + basename(p) + "](#" + anchor + ")\n"

		// read markdown from file or skip to next file
		var markdown []byte
		markdown, err = readfile(p)
		if err != nil {
			self.Logger.Error("failed to read file: %s (%s)", p, err)
			continue
		}

		// prepend anchor to content
		markdown = append([]byte("\n<a id='"+anchor+"'/>\n\n"), markdown...)

		// append a "back-to-top" anchor
		markdown = append(markdown, []byte("\n[back to top](#top)\n\n")...)

		// append to content
		content = append(content, markdown...)

		// update depth
		previous_depth = depth
	}

	// prepend toc
	content = append([]byte(toc), content...)

	// prepare output directory
	if ok, _ := exists(self.Output); !ok {
		if err = mkdirall(self.Output, 0770); err != nil {
			return err
		}
	}

	// create page object with version & content
	page := page{
		Version: self.version,
		Content: template.HTML(blackfriday.MarkdownCommon(content)),
	}

	// prepare output file path
	out := filepath.Join(self.Output, "index.html")

	// attempt to open file for output
	var f *os.File
	if f, err = create(out); err != nil {
		return err
	}
	defer f.Close()

	// prepare a writer /w buffer
	fb := bufio.NewWriter(f)
	defer fb.Flush()

	// attempt to use template to write output with page context
	err = self.template.Execute(fb, page)
	return
}

func (self *Generator) Generate() error {

	// process template
	var err error
	if self.template, err = parseFiles(self.TemplateFile); err != nil {
		return err
	}

	// acquire version
	self.version = version(self.Input)

	// sanitize input & output
	self.Input, _ = filepath.Abs(self.Input)
	self.Output, _ = filepath.Abs(self.Output)

	// sanitize & validate properties
	self.Input = filepath.Clean(self.Input)
	self.Output = filepath.Clean(self.Output)

	// walk the file system
	if err := walk(self.Input, self.walk); err != nil {
		return err
	}

	// debug: print state
	self.Logger.Debug("generator state: %+v", self)

	// determine assembly method
	if self.Book {
		return self.single()
	}
	return self.multi()
}
