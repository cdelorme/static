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

type logger interface {
	Debug(string, ...interface{})
	Error(string, ...interface{})
	Info(string, ...interface{})
}

type runner interface {
	Run(string, ...string) ([]byte, error)
}

var stat = os.Stat
var isNotExist = os.IsNotExist
var extensions = []string{".md", ".mkd", ".markdown"}
var runnable runner

type cmd struct{}

func (self cmd) Run(command string, args ...string) ([]byte, error) {
	return exec.Command(command, args...).Output()
}

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

func version(dir string) string {
	version := strconv.FormatInt(time.Now().Unix(), 10)
	out, err := runnable.Run("sh", "-c", "git", "-C", dir, "rev-parse", "--short", "HEAD")
	if err == nil {
		version = strings.Trim(string(out), "\n")
	}
	return version
}

func basename(name string) string {
	return filepath.Base(strings.TrimSuffix(name, filepath.Ext(name)))
}

func isMarkdown(path string) bool {
	for i := range extensions {
		if strings.HasSuffix(path, extensions[i]) {
			return true
		}
	}
	return false
}
