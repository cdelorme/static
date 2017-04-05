package main

import (
	"os"
	"path/filepath"

	"github.com/cdelorme/glog"
	"github.com/cdelorme/gonf"
	"github.com/cdelorme/static"

	"github.com/russross/blackfriday"
)

var exit = os.Exit
var getwd = os.Getwd
var operate = blackfriday.MarkdownCommon

func main() {
	cwd, _ := getwd()

	smd := &staticmd.Markdown{
		L:      &glog.Logger{},
		Input:  cwd,
		Output: filepath.Join(cwd, "public/"),
	}

	g := &gonf.Config{}
	g.Target(smd)
	g.Description("command line tool for generating deliverable static content")
	g.Add("template", "path to the template file", "STATICMD_TEMPLATE", "--template", "-t:")
	g.Add("input", "path to the markdown files", "STATICMD_INPUT", "--input", "-i:")
	g.Add("output", "path to place generated content", "STATICMD_OUTPUT", "--output", "-o:")
	g.Add("book", "combine all content into a single file", "STATICMD_BOOK", "--book", "-b")
	g.Add("relative", "use relative paths instead of absolute paths", "STATICMD_RELATIVE", "--relative", "-r")
	g.Example("-t template.tmpl -i . -b")
	g.Example("-t template.tmpl -i src/ -o out/ -r")
	g.Load()

	if err := smd.Generate(operate); err != nil {
		exit(1)
	}
}
