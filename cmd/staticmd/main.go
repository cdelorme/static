package main

import (
	"os"
	"path/filepath"

	"github.com/cdelorme/go-log"
	"github.com/cdelorme/go-maps"
	"github.com/cdelorme/go-option"

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

	appOptions := option.App{Description: "command line tool for generating deliverable static content"}
	appOptions.Flag("template", "path to the template file", "--template", "-t")
	appOptions.Flag("input", "path to the markdown files", "--input", "-i")
	appOptions.Flag("output", "path to place generated content", "--output", "-o")
	appOptions.Flag("book", "combine all content into a single file", "--book", "-b")
	appOptions.Flag("relative", "use relative paths instead of absolute paths", "--relative", "-r")
	appOptions.Example("-t template.tmpl -i . -b")
	appOptions.Example("-t template.tmpl -i src/ -o out/ -r")
	flags := appOptions.Parse()

	smd.TemplateFile, _ = maps.String(flags, smd.TemplateFile, "template")
	smd.Input, _ = maps.String(flags, smd.Input, "input")
	smd.Output, _ = maps.String(flags, smd.Output, "output")
	smd.Book, _ = maps.Bool(flags, smd.Book, "book")
	smd.Relative, _ = maps.Bool(flags, smd.Relative, "relative")

	return smd
}

func main() {
	smd := configure()
	if err := smd.Generate(); err != nil {
		exit(1)
	}
}
