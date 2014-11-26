package main

import (
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/cdelorme/go-log"
	"github.com/cdelorme/go-maps"
	"github.com/cdelorme/go-option"
)

// check that a path exists
// does not care if it is a directory
// will not say whether user has rw access, but
// will throw an error if the user cannot read the parent directory
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

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

// remove the path and extension from a given filename
func basename(name string) string {
	return filepath.Base(strings.TrimSuffix(name, filepath.Ext(name)))
}

func main() {

	// get current directory
	cwd, _ := os.Getwd()

	// prepare staticmd with dependencies & defaults
	staticmd := Staticmd{
		Logger:         log.Logger{Level: log.Error},
		Version:        version(),
		Input:          cwd,
		Output:         filepath.Join(cwd, "public/"),
	}

	// optimize concurrent processing
	staticmd.MaxParallelism = runtime.NumCPU()
	runtime.GOMAXPROCS(staticmd.MaxParallelism)

	// prepare cli options
	appOptions := option.App{Description: "command line tool for generating deliverable static content"}
	appOptions.Flag("template", "path to the template file", "--template", "-t")
	appOptions.Flag("input", "path to the markdown files", "--input", "-i")
	appOptions.Flag("output", "path to place generated content", "--output", "-o")
	appOptions.Flag("book", "combine all content into a single file", "--book", "-b")
	appOptions.Flag("relative", "use relative paths instead of absolute paths", "--relative", "-r")
	appOptions.Flag("debug", "verbose debug output", "--debug", "-d")
	appOptions.Flag("profile", "produce profile output to supplied path", "--profile", "-p")
	appOptions.Example("-t template.tmpl -i . -b")
	appOptions.Example("-t template.tmpl -i src/ -o out/ -r")
	flags := appOptions.Parse()

	// apply flags
	t, _ := maps.String(&flags, "", "template")
	if tmpl, err := template.ParseFiles(t); err != nil {
		staticmd.Logger.Error("Failed to open template: %s", err)
		os.Exit(1)
	} else {
		staticmd.Template = *tmpl
	}
	staticmd.Input, _ = maps.String(&flags, staticmd.Input, "input")
	staticmd.Output, _ = maps.String(&flags, staticmd.Output, "output")
	staticmd.Book, _ = maps.Bool(&flags, staticmd.Book, "book")
	staticmd.Relative, _ = maps.Bool(&flags, staticmd.Relative, "relative")

	// sanitize input & output
	staticmd.Input, _ = filepath.Abs(staticmd.Input)
	staticmd.Output, _ = filepath.Abs(staticmd.Output)

	// optionally enable debugging
	if debug, _ := maps.Bool(&flags, false, "debug"); debug {
		staticmd.Logger.Level = log.Debug
	}

	// optionally enable profiling
	if profile, _ := maps.String(&flags, "", "profile"); profile != "" {
		f, _ := os.Create(profile)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// sanitize & validate properties
	staticmd.Input = filepath.Clean(staticmd.Input)
	staticmd.Output = filepath.Clean(staticmd.Output)

	// print debug status
	staticmd.Logger.Debug("Staticmd State: %+v", staticmd)

	// walk the file system
	if err := filepath.Walk(staticmd.Input, staticmd.Walk); err != nil {
		staticmd.Logger.Error("failed to walk directory: %s", err)
	}
	staticmd.Logger.Debug("Pages: %+v", staticmd.Pages)

	// build
	if staticmd.Book {
		staticmd.Single()
	} else {
		staticmd.Multi()
	}
}
