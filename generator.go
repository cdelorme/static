package staticmd

import (
	"bufio"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/russross/blackfriday"
)

var MaxParallelism = runtime.NumCPU()

type Generator struct {
	Input        string
	Output       string
	TemplateFile string
	Book         bool
	Relative     bool
	Logger       logger

	version  string
	pages    []string
	template *template.Template
}

// convert markdown input path to html output path
// there is no reverse (because we support `.md`, `.mkd`, and `.markdown`)
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

// walk the directories and build a list of pages
func (self *Generator) walk(path string, file os.FileInfo, err error) error {

	// only pay attention to files with a size greater than 0
	if file == nil {
	} else if file.Mode().IsRegular() && file.Size() > 0 {

		// only add markdown files to our pages array (.md, .mkd, .markdown)
		if strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".mkd") || strings.HasSuffix(path, ".markdown") {
			self.pages = append(self.pages, path)
		}
	}
	return err
}

// build output concurrently to many pages
func (self *Generator) multi() error {

	// prepare navigation storage
	navigation := make(map[string][]Navigation)

	// loop pages to build table of contents
	for i, _ := range self.pages {

		// build output directory
		out := self.ior(self.pages[i])
		dir := filepath.Dir(self.ior(out))

		// create navigation object
		nav := Navigation{}

		// sub-index condition changes name, dir, and link
		if filepath.Dir(out) != self.Output && strings.ToLower(basename(out)) == "index" {

			// set name to containing folder
			nav.Name = basename(dir)

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
			nav.Name = basename(out)

			// set relative or absolute link
			if self.Relative {
				nav.Link = strings.TrimPrefix(out, filepath.Dir(out)+string(os.PathSeparator))
			} else {
				nav.Link = strings.TrimPrefix(out, self.Output)
			}
		}

		// build indexes first-match
		if _, ok := navigation[dir]; !ok {
			navigation[dir] = make([]Navigation, 0)

			// create output directory for when we create files
			if ok, _ := exists(dir); !ok {
				if err := os.MkdirAll(dir, 0770); err != nil {
					self.Logger.Error("Failed to create path: %s, %s", dir, err)
				}
			}
		}

		// append to navigational list
		navigation[dir] = append(navigation[dir], nav)
	}

	// debug navigation output
	self.Logger.Debug("navigation: %+v", navigation)

	// prepare waitgroup, bufferer channel, and add number of async handlers to wg
	var wg sync.WaitGroup
	pages := make(chan string, MaxParallelism)
	wg.Add(MaxParallelism)

	// prepare workers
	for i := 0; i < MaxParallelism; i++ {
		go func() {
			defer wg.Done()

			// iterate supplied pages
			for p := range pages {

				// acquire output filepath
				out := self.ior(p)
				dir := filepath.Dir(out)

				// prepare a new page object for our template to render
				page := Page{
					Name:    basename(p),
					Version: self.version,
					Nav:     navigation[self.Output],
					Depth:   self.depth(out),
				}

				// read in page text
				if markdown, err := ioutil.ReadFile(p); err == nil {

					// if this page happens to be a sub-index we can generate the table of contents
					if dir != self.Output && strings.ToLower(basename(p)) == "index" {

						// iterate and build table of contents as basic markdown
						toc := "\n## Table of Contents:\n\n"
						for i, _ := range navigation[dir] {
							toc = toc + "- [" + navigation[dir][i].Name + "](" + navigation[dir][i].Link + ")\n"
						}

						// debug table of contents output
						self.Logger.Debug("table of contents for %s, %s", out, toc)

						// prepend table of contents
						markdown = append([]byte(toc), markdown...)
					}

					// convert to html, and accept as part of the template
					page.Content = template.HTML(blackfriday.MarkdownCommon(markdown))
				} else {
					self.Logger.Error("failed to read file: %s, %s", p, err)
				}

				// debug page output
				self.Logger.Debug("page: %+v", page)

				// translate input path to output path & create a write context
				if f, err := os.Create(out); err == nil {
					defer f.Close()

					// prepare a writer /w buffer
					fb := bufio.NewWriter(f)
					defer fb.Flush()

					// attempt to use template to write output with page context
					if e := self.template.Execute(fb, page); e != nil {
						self.Logger.Error("Failed to write template: %s, %s", out, e)
					}
				} else {
					self.Logger.Error("failed to create new file: %s, %s", out, err)
				}
			}
		}()
	}

	// send pages to workers for async rendering
	for i, _ := range self.pages {
		pages <- self.pages[i]
	}

	// close channel and wait for async to finish before continuing
	close(pages)
	wg.Wait()

	return nil
}

// build output synchronously to a single page
func (self *Generator) single() error {

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
		markdown, err := ioutil.ReadFile(p)
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

	// create page object with version & content
	page := Page{
		Version: self.version,
		Content: template.HTML(blackfriday.MarkdownCommon(content)),
	}

	// prepare output directory
	if ok, _ := exists(self.Output); !ok {
		if err := os.MkdirAll(self.Output, 0770); err != nil {
			return err
		}
	}

	// prepare output file path
	out := filepath.Join(self.Output, "index.html")

	// attempt to open file for output
	var f *os.File
	var err error
	if f, err = os.Create(out); err != nil {
		return err
	}
	defer f.Close()

	// prepare a writer /w buffer
	fb := bufio.NewWriter(f)
	defer fb.Flush()

	// attempt to use template to write output with page context
	return self.template.Execute(fb, page)
}

func (self *Generator) Generate() error {

	// process template
	var err error
	if self.template, err = template.ParseFiles(self.TemplateFile); err != nil {
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

	// debug: print state
	self.Logger.Debug("generator state: %+v", self)

	// walk the file system
	if err := filepath.Walk(self.Input, self.walk); err != nil {
		return err
	}

	// debug: print pages
	self.Logger.Debug("pages: %+v", self.pages)

	// determine assembly method
	if self.Book {
		return self.single()
	}
	return self.multi()
}
