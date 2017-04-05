package static

import (
	"html/template"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var parseTemplate *template.Template
var readfileError, templateError, createError, mkdirallError, parseError, walkError error

func init() {
	readfile = func(_ string) ([]byte, error) { return nil, readfileError }
	create = func(_ string) (*os.File, error) { return nil, createError }
	mkdirall = func(_ string, _ os.FileMode) error { return mkdirallError }
	parseFiles = func(...string) (*template.Template, error) { return parseTemplate, parseError }
	walk = func(_ string, _ filepath.WalkFunc) error { return walkError }
}

type mockLogger struct{}

func (self *mockLogger) Error(_ string, _ ...interface{}) {}
func (self *mockLogger) Debug(_ string, _ ...interface{}) {}
func (self *mockLogger) Info(_ string, _ ...interface{})  {}

type mockTemplate struct{}

func (self *mockTemplate) Execute(_ io.Writer, _ interface{}) error { return templateError }

type mockFileInfo struct {
	N  string
	S  int64
	Fm uint32
	T  time.Time
	D  bool
	So interface{}
}

func (self *mockFileInfo) Name() string       { return self.N }
func (self *mockFileInfo) Size() int64        { return self.S }
func (self *mockFileInfo) Mode() os.FileMode  { return os.FileMode(self.Fm) }
func (self *mockFileInfo) ModTime() time.Time { return self.T }
func (self *mockFileInfo) IsDir() bool        { return self.D }
func (self *mockFileInfo) Sys() interface{}   { return self.So }

func TestIor(t *testing.T) {
	t.Parallel()
	g := Markdown{}
	if s := g.ior("some.md"); len(s) == 0 {
		t.FailNow()
	}
}

func TestDepth(t *testing.T) {
	t.Parallel()
	absp := "/abs/path/"
	g := Markdown{Output: absp}

	// test abs depth
	if d := g.depth("somefile"); len(d) > 0 {
		t.FailNow()
	}

	// test relative depth
	g.Relative = true
	if d := g.depth(absp + "somefile"); len(d) == 0 {
		t.Logf("Path: %s\n", d)
		t.FailNow()
	}
}

func TestWalk(t *testing.T) {
	t.Parallel()
	g := Markdown{}

	p := "valid.md"
	var f os.FileInfo = &mockFileInfo{S: 1}
	var e error

	// test with valid file
	if err := g.walk(p, f, e); err != nil {
		t.FailNow()
	}
}

func TestMulti(t *testing.T) {
	g := Markdown{L: &mockLogger{}, template: &mockTemplate{}, pages: []string{"fuck.md", "deeper/than/index.md", "deeper/than/data.md"}}

	// set expected defaults
	notExist = false
	statError = nil

	// no pages
	if e := g.multi(); e != nil {
		t.FailNow()
	}

	// test full pass
	if e := g.multi(); e != nil {
		t.FailNow()
	}

	// test full pass relative
	g.Relative = true
	if e := g.multi(); e != nil {
		t.FailNow()
	}

	// test failing execute
	templateError = mockError
	if e := g.multi(); e == nil {
		t.FailNow()
	}

	// test failing file creation
	createError = mockError
	if e := g.multi(); e == nil {
		t.FailNow()
	}

	// test failing to read the file
	readfileError = mockError
	if e := g.multi(); e == nil {
		t.FailNow()
	}

	// test dir creation failure
	mkdirallError = mockError
	statError = mockError
	if e := g.multi(); e == nil {
		t.FailNow()
	}
}

func TestSingle(t *testing.T) {
	g := Markdown{L: &mockLogger{}, template: &mockTemplate{}, pages: []string{"fuck.md", "deeper/than/index.md", "deeper/than/data.md"}}

	// reset expected defaults
	statError = nil
	readfileError = nil
	createError = nil
	templateError = nil

	// test full pass
	if e := g.single(); e != nil {
		t.FailNow()
	}

	// test failing execute
	templateError = mockError
	if e := g.single(); e == nil {
		t.FailNow()
	}

	// test create error
	createError = mockError
	if e := g.single(); e == nil {
		t.FailNow()
	}

	// test fail mkdirall
	mkdirallError = mockError
	statError = mockError
	if e := g.single(); e == nil {
		t.FailNow()
	}

	// test fail readfile
	readfileError = mockError
	if e := g.single(); e == nil {
		t.FailNow()
	}
}

func TestGenerate(t *testing.T) {
	g := Markdown{L: &mockLogger{}}

	// set template for stand-alone execution
	parseTemplate = template.New("test")

	// test full pass
	if e := g.Generate(); e != nil {
		t.FailNow()
	}

	// test book mode full pass
	g.Book = true
	if e := g.Generate(); e == nil {
		t.FailNow()
	}

	// test walk error
	walkError = mockError
	if e := g.Generate(); e == nil {
		t.FailNow()
	}

	// test template error
	parseError = mockError
	if e := g.Generate(); e == nil {
		t.FailNow()
	}
}
