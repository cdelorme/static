package main

import (
	"bufio"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/russross/blackfriday"

	"github.com/cdelorme/go-log"
)

type Staticmd struct {
	Logger         log.Logger
	Input          string
	Output         string
	Template       template.Template
	Book           bool
	Relative       bool
	MaxParallelism int
	Version        string
	Pages          []string
}

// convert markdown input path to html output path
// there is no reverse (because we support `.md`, `.mkd`, and `.markdown`)
func (staticmd *Staticmd) ior(path string) string {
	return strings.TrimSuffix(strings.Replace(path, staticmd.Input, staticmd.Output, 1), filepath.Ext(path)) + ".html"
}

// for relative rendering generate relative depth string, or fallback to empty
func (staticmd *Staticmd) depth(path string) string {
	if staticmd.Relative {
		if rel, err := filepath.Rel(filepath.Dir(path), staticmd.Output); err == nil {
			return rel+string(os.PathSeparator)
		}
	}
	return ""
}

// walk the directories and build a list of pages
func (staticmd *Staticmd) Walk(path string, file os.FileInfo, err error) error {

	// only pay attention to files with a size greater than 0
	if file == nil {
	} else if file.Mode().IsRegular() && file.Size() > 0 {

		// only add markdown files to our pages array (.md, .mkd, .markdown)
		if strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".mkd") || strings.HasSuffix(path, ".markdown") {
			staticmd.Pages = append(staticmd.Pages, path)
		}
	}
	return err
}

// build output concurrently to many pages
func (staticmd *Staticmd) Multi() {

	// prepare navigation storage
	navigation := make(map[string][]Navigation)

	// loop pages to build table of contents
	for i, _ := range staticmd.Pages {

		// build output directory
		out := staticmd.ior(staticmd.Pages[i])
		dir := filepath.Dir(staticmd.ior(out))

		// create navigation object
		nav := Navigation{}

		// sub-index condition changes name, dir, and link
		if filepath.Dir(out) != staticmd.Output && strings.ToLower(basename(out)) == "index" {

			// set name to containing folder
			nav.Name = basename(dir)

			// set relative or absolute link
			if staticmd.Relative {
                nav.Link = filepath.Join(strings.TrimPrefix(dir, filepath.Dir(dir)+string(os.PathSeparator)), filepath.Base(out))
			} else {
				nav.Link = strings.TrimPrefix(dir, staticmd.Output)+string(os.PathSeparator)
			}

			// update dir to dir of dir
			dir = filepath.Dir(dir)
		} else {

			// set name to files name
			nav.Name = basename(out)

			// set relative or absolute link
			if staticmd.Relative {
				nav.Link = strings.TrimPrefix(out, filepath.Dir(out)+string(os.PathSeparator))
			} else {
				nav.Link = strings.TrimPrefix(out, staticmd.Output)
			}
		}

		// build indexes first-match
		if _, ok := navigation[dir]; !ok {
			navigation[dir] = make([]Navigation, 0)

			// create output directory for when we create files
			if ok, _ := exists(dir); !ok {
				if err := os.MkdirAll(dir, 0770); err != nil {
					staticmd.Logger.Error("Failed to create path: %s, %s", dir, err)
				}
			}
		}

		// append to navigational list
		navigation[dir] = append(navigation[dir], nav)
	}

	// debug navigation output
	staticmd.Logger.Debug("navigation: %+v", navigation)

	// prepare waitgroup, bufferer channel, and add number of async handlers to wg
	var wg sync.WaitGroup
	pages := make(chan string, staticmd.MaxParallelism)
	wg.Add(staticmd.MaxParallelism)

	// prepare workers
	for i := 0; i < staticmd.MaxParallelism; i++ {
		go func() {
			defer wg.Done()

			// iterate supplied pages
			for p := range pages {

				// acquire output filepath
				out := staticmd.ior(p)
				dir := filepath.Dir(out)

				// prepare a new page object for our template to render
				page := Page{
					Name:    basename(p),
					Version: staticmd.Version,
					Nav:     navigation[staticmd.Output],
					Depth:   staticmd.depth(out),
				}

				// read in page text
				if markdown, err := ioutil.ReadFile(p); err == nil {

					// if this page happens to be a sub-index we can generate the table of contents
					if dir != staticmd.Output && strings.ToLower(basename(p)) == "index" {

						// iterate and build table of contents as basic markdown
						toc := "\n## Table of Contents:\n\n"
						for i, _ := range navigation[dir] {
							toc = toc + "- [" + navigation[dir][i].Name + "](" + navigation[dir][i].Link + ")\n"
						}

						// debug table of contents output
						staticmd.Logger.Debug("table of contents for %s, %s", out, toc)

						// prepend table of contents
						markdown = append([]byte(toc), markdown...)
					}

					page.Content = template.HTML(blackfriday.MarkdownCommon(markdown))
				} else {
					staticmd.Logger.Error("failed to read file: %s, %s", p, err)
				}

				// debug page output
				staticmd.Logger.Debug("page: %+v", page)

				// translate input path to output path & create a write context
				if f, err := os.Create(out); err == nil {
					defer f.Close()

					// prepare a writer /w buffer
					fb := bufio.NewWriter(f)
					defer fb.Flush()

					// attempt to use template to write output with page context
					if e := staticmd.Template.Execute(fb, page); e != nil {
						staticmd.Logger.Error("Failed to write template: %s, %s", out, e)
					}
				} else {
					staticmd.Logger.Error("failed to create new file: %s, %s", out, err)
				}
			}
		}()
	}

	// send pages to workers for async rendering
	for _, page := range staticmd.Pages {
		pages <- page
	}

	// close channel and wait for async to finish before continuing
	close(pages)
	wg.Wait()
}

// build output synchronously to a single page
func (staticmd *Staticmd) Single() {

	// still determining strategy

}
