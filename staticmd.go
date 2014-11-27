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
func (staticmd *Staticmd) ior(path string) string {
	return strings.TrimSuffix(strings.Replace(path, staticmd.Input, staticmd.Output, 1), filepath.Ext(path)) + ".html"
}

// for relative rendering generate relative depth string, or fallback to empty
func (staticmd *Staticmd) depth(path string) string {
	if staticmd.Relative {
		if rel, err := filepath.Rel(filepath.Dir(path), staticmd.Output); err == nil {
			return rel
		}
	}
	return ""
}

// get link to file, with support for relative path linking
func (staticmd *Staticmd) link(path string) string {
	if staticmd.Relative {
		return strings.TrimPrefix(path, filepath.Dir(path))
	}
	return strings.TrimPrefix(path, staticmd.Output)
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

		// create a new navigation object
		nav := Navigation{
			Name: basename(out),
			Link: staticmd.link(out),
		}

		// append files to their respective directories
		navigation[dir] = append(navigation[dir], nav)
	}

	// @todo second cycle to clarify whether index or readme for table-of-contents

	// debug output
	staticmd.Logger.Debug("Navigation: %+v", navigation)

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
				// dir := filepath.Dir(out)

				// prepare a new page object for our template to render
				page := Page{
					Name:    basename(p),
					Version: staticmd.Version,
					Nav:     navigation[staticmd.Output],
					Depth:   staticmd.depth(p),
				}

				// read in page text
				if markdown, err := ioutil.ReadFile(p); err == nil {

					// conditionally prepend table of contents?

					page.Content = template.HTML(blackfriday.MarkdownCommon(markdown))
				} else {
					staticmd.Logger.Error("failed to read file: %s, %s", p, err)
				}

				// translate input path to output path & create a write context
				if f, err := os.Create(out); err == nil {
					fb := bufio.NewWriter(f)
					defer fb.Flush()
					defer f.Close()

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
