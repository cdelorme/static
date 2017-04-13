package static

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
)

type mockLogger struct{}

func (l *mockLogger) Debug(string, ...interface{}) {}
func (l *mockLogger) Info(string, ...interface{})  {}
func (l *mockLogger) Error(string, ...interface{}) {}

func TestMarkdown(t *testing.T) {
	var files []*os.File

	// abstract behaviors
	o := func(b []byte) []byte { return b }
	readall = func(f io.Reader) ([]byte, error) { return []byte{}, nil }
	mkdirall = func(d string, f os.FileMode) error { return nil }
	create = func(n string) (*os.File, error) {
		tfo, e := ioutil.TempFile(os.TempDir(), "static-out")
		files = append(files, tfo)
		return tfo, e
	}
	open = func(n string) (*os.File, error) {
		tfi, e := ioutil.TempFile(os.TempDir(), "static-in")
		files = append(files, tfi)
		return tfi, e
	}

	// execute operation book mode
	m := &Markdown{L: &mockLogger{}}
	if e := m.Run(o); e != nil {
		t.Error(e)
	}

	// execute operation web mode
	m.Web = true
	if e := m.Run(o); e != nil {
		t.Error(e)
	}

	// remove all temporary files
	for i := range files {
		os.Remove(files[i].Name())
	}
}
