package staticmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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

type logger interface {
	Debug(string, ...interface{})
	Error(string, ...interface{})
	Info(string, ...interface{})
}
