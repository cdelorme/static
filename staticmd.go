package main

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"

	// "bufio"
	// "io/ioutil"
	// "path"

	// "github.com/russross/blackfriday"

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
	Navigation     []string
	Pages          []string
	Subdirectories map[string][]string
	Indexes        map[string][]string
}

// search for readme or fallback to index
func (staticmd *Staticmd) index(pages []string) string {
	for _, page := range pages {
		if n := basename(page); n == "readme" {
			return n
		}
	}
	return "index"
}

func (staticmd *Staticmd) Walk(path string, file os.FileInfo, err error) error {
	if file == nil {
	} else if file.IsDir() {

		// prepare path to create for matching output
		mkd := filepath.Join(staticmd.Output, strings.TrimPrefix(path, staticmd.Input))

		// include parent output path
		if path == staticmd.Input {
			mkd = staticmd.Output
		} else {

			// note subdirectory for indexing
			if _, ok := staticmd.Subdirectories[filepath.Dir(path)]; !ok {
				staticmd.Subdirectories[filepath.Dir(path)] = make([]string, 0)
			}
			staticmd.Subdirectories[filepath.Dir(path)] = append(staticmd.Subdirectories[filepath.Dir(path)], path)
		}

		// create all matching output paths
		staticmd.Logger.Debug("creating matching output path: %s", mkd)
		if err := os.MkdirAll(mkd, 0770); err != nil {
			staticmd.Logger.Error("failed to create matching output path: %s, %s", path, err)
		}

	} else if file.Mode().IsRegular() && file.Size() > 0 {

		// append all markdown files to our pages list
		if strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".mkd") || strings.HasSuffix(path, ".markdown") {
			staticmd.Pages = append(staticmd.Pages, path)

			dir := filepath.Dir(path)

			// append to navigation, index
			if dir == staticmd.Input {
				staticmd.Navigation = append(staticmd.Navigation, path)
			} else {

				// prepare index container
				if _, ok := staticmd.Indexes[dir]; !ok {
					staticmd.Indexes[dir] = make([]string, 0)
				}
				staticmd.Indexes[dir] = append(staticmd.Indexes[dir], dir+basename(strings.TrimPrefix(path, staticmd.Input))+".html")
			}
		}
	}
	return err
}

func (staticmd *Staticmd) Multi() {

	// test filepath.Rel()

	// iterate all subdirectories to find index/readme and append to
	// for i, _ := range staticmd.Subdirectories {
	// 	for _, d := range staticmd.Subdirectories[i] {

	// 		// can only append to indexes that exist
	// 		if _, ok := staticmd.Indexes[filepath.Dir(staticmd.Subdirectories[i][d])]; !ok {
	// 			continue
	// 		}

	// 		// determine whether to append relative path or just the folder
	// 		if staticmd.Relative {

	// 			// trim input path, and append to indexes as .html
	// 			idx := staticmd.index(staticmd.Subdirectories[i][d])
	// 			staticmd.Subdirectories[i][d]+idx+".html"

	// 			// determine whether index or readme exists, assume index
	// 		} else {
	// 			// simply append `subfolder/`
	// 		}
	// 	}
	// }

	// rebuild navigation pathing
	// for i, page := range staticmd.Navigation {
	// 	if staticmd.Relative {
	// 		staticmd.Navigation[i] = strings.TrimPrefix(strings.TrimPrefix(page, staticmd.Input), "/")
	// 	} else if basename(page) == "index" {
	// 		staticmd.Navigation[i] = "/"
	// 	} else {
	// 		staticmd.Navigation[i] = strings.TrimPrefix(page, staticmd.Input)
	// 	}
	// }

	// debug output
	staticmd.Logger.Debug("Navigation: %+v", staticmd.Navigation)
	staticmd.Logger.Debug("Indexes: %+v", staticmd.Indexes)

	// concurrently build pages

}

func (staticmd *Staticmd) Single() {

	// open single index.html for building
	file := staticmd.Output

	// prepare index
	index := make([]string, 0)

	// prepare markdown thingy?
	// content := template.HTML(blackfriday.MarkdownCommon(markdown))

	// synchronously build one output file
	for i, _ := range staticmd.Pages {

		// create index record

		// staticmd.Pages[i]
		// append each file to content
	}

	// append index

}
