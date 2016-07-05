package main

import (
	"html/template"
	"os"
	"path/filepath"
	"runtime/pprof"

	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/cdelorme/go-log"
	"github.com/cdelorme/go-maps"
	"github.com/cdelorme/go-option"

	"github.com/cdelorme/staticmd"
)

// if within a git repo, gets git version as a short-hash
// otherwise falls back to a unix timestamp
func version() string {
	version := strconv.FormatInt(time.Now().Unix(), 10)
	out, err := exec.Command("sh", "-c", "git rev-parse --short HEAD").Output()
	if err == nil {
		version = strings.Trim(string(out), "\n")
	}
	return version
}

func main() {

	// get current directory
	cwd, _ := os.Getwd()

	// prepare staticmd with dependencies & defaults
	gen := staticmd.Staticmd{
		Logger:  log.Logger{},
		Version: version(),
		Input:   cwd,
		Output:  filepath.Join(cwd, "public/"),
	}

	// prepare cli options
	appOptions := option.App{Description: "command line tool for generating deliverable static content"}
	appOptions.Flag("template", "path to the template file", "--template", "-t")
	appOptions.Flag("input", "path to the markdown files", "--input", "-i")
	appOptions.Flag("output", "path to place generated content", "--output", "-o")
	appOptions.Flag("book", "combine all content into a single file", "--book", "-b")
	appOptions.Flag("relative", "use relative paths instead of absolute paths", "--relative", "-r")
	appOptions.Flag("profile", "produce profile output to supplied path", "--profile", "-p")
	appOptions.Example("-t template.tmpl -i . -b")
	appOptions.Example("-t template.tmpl -i src/ -o out/ -r")
	flags := appOptions.Parse()

	// apply flags
	t, _ := maps.String(flags, "", "template")
	if tmpl, err := template.ParseFiles(t); err != nil {
		gen.Logger.Error("Failed to open template: %s", err)
		os.Exit(1)
	} else {
		gen.Template = *tmpl
	}
	gen.Input, _ = maps.String(flags, gen.Input, "input")
	gen.Output, _ = maps.String(flags, gen.Output, "output")
	gen.Book, _ = maps.Bool(flags, gen.Book, "book")
	gen.Relative, _ = maps.Bool(flags, gen.Relative, "relative")

	// sanitize input & output
	gen.Input, _ = filepath.Abs(gen.Input)
	gen.Output, _ = filepath.Abs(gen.Output)

	// optionally enable profiling
	if profile, _ := maps.String(flags, "", "profile"); profile != "" {
		f, _ := os.Create(profile)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// sanitize & validate properties
	gen.Input = filepath.Clean(gen.Input)
	gen.Output = filepath.Clean(gen.Output)

	// print debug status
	gen.Logger.Debug("Staticmd State: %+v", gen)

	// walk the file system
	if err := filepath.Walk(gen.Input, gen.Walk); err != nil {
		gen.Logger.Error("failed to walk directory: %s", err)
	}
	gen.Logger.Debug("Pages: %+v", gen.Pages)

	// build
	if gen.Book {
		gen.Single()
	} else {
		gen.Multi()
	}
}
