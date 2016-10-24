package main

import (
	"os"
	"path/filepath"

	"github.com/cdelorme/go-log"
	"github.com/cdelorme/gonf"
	"github.com/cdelorme/staticmd"
)

var exit = os.Exit
var getwd = os.Getwd

type generator interface {
	Generate() error
}

func configure() generator {
	cwd, _ := getwd()

	smd := &staticmd.Generator{
		Logger: &log.Logger{},
		Input:  cwd,
		Output: filepath.Join(cwd, "public/"),
	}

	g := &gonf.Gonf{Description: "command line tool for generating deliverable static content", Configuration: smd}
	g.Add("template", "path to the template file", "STATICMD_TEMPLATE", "--template", "-t:")
	g.Add("input", "path to the markdown files", "STATICMD_INPUT", "--input", "-i:")
	g.Add("output", "path to place generated content", "STATICMD_OUTPUT", "--output", "-o:")
	g.Add("book", "combine all content into a single file", "STATICMD_BOOK", "--book", "-b")
	g.Add("relative", "use relative paths instead of absolute paths", "STATICMD_RELATIVE", "--relative", "-r")
	g.Example("-t template.tmpl -i . -b")
	g.Example("-t template.tmpl -i src/ -o out/ -r")
	g.Load()

	return smd
}

func main() {
	smd := configure()
	if err := smd.Generate(); err != nil {
		exit(1)
	}
}
