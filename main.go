package main

import (
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
func Version() string {
	version := strconv.FormatInt(time.Now().Unix(), 10)
	out, err := exec.Command("sh", "-c", "git rev-parse --short HEAD").Output()
	if err == nil {
		version = strings.Trim(string(out), "\n")
	}
	return version
}

func main() {

	// prepare staticmd with dependencies
	staticmd := Staticmd{
		Logger: log.Logger{Level: log.Error},
		Subdirectories: make(map[string][]string),
		Indexes:        make(map[string][]string),
	}

	// optimize concurrent processing
	staticmd.MaxParallelism = runtime.NumCPU()
	runtime.GOMAXPROCS(staticmd.MaxParallelism)

	// get current directory
	cwd, _ := os.Getwd()

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
	staticmd.Template, _ = maps.String(&flags, staticmd.Template, "template")
	staticmd.Input, _ = maps.String(&flags, cwd, "input")
	staticmd.Output, _ = maps.String(&flags, filepath.Join(cwd, "public/"), "output")
	staticmd.Book, _ = maps.Bool(&flags, staticmd.Book, "book")

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

	if err := filepath.Walk(staticmd.Input, staticmd.Walk); err != nil {
		staticmd.Logger.Error("failed to walk directory: %s", err)
	}
	staticmd.Logger.Debug("Pages: %+v", staticmd.Pages)

	// build indexes (includes navigation)
	staticmd.Index()
	staticmd.Logger.Debug("Navigation: %+v", staticmd.Navigation)
	staticmd.Logger.Debug("Indexes: %+v", staticmd.Indexes)

	// parse files
	if staticmd.Book {
		staticmd.BuildSingle()
	} else {
		staticmd.BuildMulti()
	}
}
