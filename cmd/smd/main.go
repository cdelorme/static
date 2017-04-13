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

	smd := &static.Markdown{
		L:      &glog.Logger{},
		Input:  cwd,
		Output: filepath.Join(cwd, "public/"),
	}

	g := &gonf.Config{}
	g.Target(smd)
	g.Description("command line tool for generating static html from markdown")
	g.Add("web", "parse into individual files matching the original file name", "STATIC_WEB", "--web", "-w")
	g.Add("title", "the title to give to the processed files", "STATIC_TITLE", "--title", "-t:")
	g.Add("input", "path to the markdown files", "STATIC_INPUT", "--input", "-i:")
	g.Add("output", "path to place generated content", "STATIC_OUTPUT", "--output", "-o:")
	g.Add("version", "optional user-defined version", "STATIC_VERSION", "--version", "-v:")
	g.Add("template", "path to user-defined template file", "STATIC_TEMPLATE", "--template")
	g.Example("-t template.tmpl -i . -b")
	g.Example("-t template.tmpl -i src/ -o out/ -r")
	g.Load()

	if err := smd.Run(operate); err != nil {
		exit(1)
	}
}
