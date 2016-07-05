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

type logger interface {
	Error(string, ...interface{})
}

func configure() (generator, logger) {

	// get current directory
	cwd, _ := getwd()

	// prepare staticmd with dependencies & defaults
	smd := &staticmd.Generator{
		Logger: &log.Logger{},
		Input:  cwd,
		Output: filepath.Join(cwd, "public/"),
	}

	// prepare cli options
	appOptions := option.App{Description: "command line tool for generating deliverable static content"}
	appOptions.Flag("template", "path to the template file", "--template", "-t")
	appOptions.Flag("input", "path to the markdown files", "--input", "-i")
	appOptions.Flag("output", "path to place generated content", "--output", "-o")
	appOptions.Flag("book", "combine all content into a single file", "--book", "-b")
	appOptions.Flag("relative", "use relative paths instead of absolute paths", "--relative", "-r")
	appOptions.Example("-t template.tmpl -i . -b")
	appOptions.Example("-t template.tmpl -i src/ -o out/ -r")
	flags := appOptions.Parse()

	// apply flags
	smd.TemplateFile, _ = maps.String(flags, smd.TemplateFile, "template")
	smd.Input, _ = maps.String(flags, smd.Input, "input")
	smd.Output, _ = maps.String(flags, smd.Output, "output")
	smd.Book, _ = maps.Bool(flags, smd.Book, "book")
	smd.Relative, _ = maps.Bool(flags, smd.Relative, "relative")

	return smd, smd.Logger
}

func main() {
	smd, l := configure()
	if err := smd.Generate(); err != nil {
		l.Error("generator failed (%s)", err)
		exit(1)
	}
}
