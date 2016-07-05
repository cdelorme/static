package staticmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func init() {
	runnable = cmd{}
}

// logger component for printing status messages to stderr
type logger interface {
	Debug(string, ...interface{})
	Error(string, ...interface{})
	Info(string, ...interface{})
}

var stat = os.Stat
var isNotExist = os.IsNotExist

type runner interface {
	Run(string, ...string) ([]byte, error)
}

var runnable runner

type cmd struct{}

func (self cmd) Run(command string, args ...string) ([]byte, error) {
	return exec.Command(command, args...).Output()
}

// check that a path exists
// does not care if it is a directory
// will not say whether user has rw access, but
// will throw an error if the user cannot read the parent directory
func exists(path string) (bool, error) {
	_, err := stat(path)
	if err == nil {
		return true, nil
	}
	if isNotExist(err) {
		return false, nil
	}
	return false, err
}

// if within a git repo, gets git version as a short-hash
// otherwise falls back to a unix timestamp
func version(dir string) string {
	version := strconv.FormatInt(time.Now().Unix(), 10)
	out, err := runnable.Run("sh", "-c", "git", "-C", dir, "rev-parse", "--short", "HEAD")
	if err == nil {
		version = strings.Trim(string(out), "\n")
	}
	return version
}

// remove the path and extension from a given filename
func basename(name string) string {
	return filepath.Base(strings.TrimSuffix(name, filepath.Ext(name)))
}
