package staticmd

import (
	"html/template"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var mockReadfileBytes []byte
var mockReadfileError error
var mockTemplateError error
var mockCreateFile *os.File
var mockCreateError error
var mockMkdirError error
var mockParseTemplate *template.Template
var mockParseError error
var mockWalkError error

func init() {
	readfile = func(_ string) ([]byte, error) { return mockReadfileBytes, mockReadfileError }
	create = func(_ string) (*os.File, error) { return mockCreateFile, mockCreateError }
	mkdirall = func(_ string, _ os.FileMode) error { return mockMkdirError }
	parseFiles = func(...string) (*template.Template, error) { return mockParseTemplate, mockParseError }
	walk = func(_ string, _ filepath.WalkFunc) error { return mockWalkError }
}

type mockLogger struct{}

func (self *mockLogger) Error(_ string, _ ...interface{}) {}
func (self *mockLogger) Debug(_ string, _ ...interface{}) {}
func (self *mockLogger) Info(_ string, _ ...interface{})  {}

type mockTemplate struct{}

func (self *mockTemplate) Execute(_ io.Writer, _ interface{}) error { return mockTemplateError }

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
	g := Generator{}
	if s := g.ior("some.md"); len(s) == 0 {
		t.FailNow()
	}
}

func TestDepth(t *testing.T) {
	t.Parallel()
	absp := "/abs/path/"
	g := Generator{Output: absp}

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
	g := Generator{}

	p := "valid.md"
	var f os.FileInfo = &mockFileInfo{S: 1}
	var e error

	// test with valid file
	if err := g.walk(p, f, e); err != nil {
		t.FailNow()
	}
}

func TestMulti(t *testing.T) {
	g := Generator{Logger: &mockLogger{}, template: &mockTemplate{}, pages: []string{"fuck.md", "deeper/than/index.md", "deeper/than/data.md"}}

	// reset defaults for parameters
	mockCreateError = nil
	mockReadfileError = nil
	mockMkdirError = nil
	statError = nil
	notExist = false
	mockCreateFile = &os.File{}

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

	// test failing file creation
	mockCreateError = mockError
	if e := g.multi(); e == nil {
		t.FailNow()
	}

	// test failing to read the file
	mockReadfileError = mockError
	if e := g.multi(); e == nil {
		t.FailNow()
	}

	// test dir creation failure
	mockMkdirError = mockError
	statError = mockError
	notExist = true
	if e := g.multi(); e == nil {
		t.FailNow()
	}
}

func TestSingle(t *testing.T) {
	t.Parallel()
	g := Generator{Logger: &mockLogger{}, template: &mockTemplate{}, pages: []string{"fuck.md", "deeper/than/index.md", "deeper/than/data.md"}}

	// reset defaults for parameters
	mockCreateError = nil
	mockReadfileError = nil
	mockMkdirError = nil
	statError = nil
	notExist = false
	mockCreateFile = &os.File{}

	// test full pass
	if e := g.single(); e != nil {
		t.FailNow()
	}

	// test create error
	mockCreateError = mockError
	if e := g.single(); e == nil {
		t.FailNow()
	}

	// test fail mkdirall
	mockMkdirError = mockError
	statError = mockError
	if e := g.single(); e == nil {
		t.FailNow()
	}

	// test fail readfile
	mockReadfileError = mockError
	if e := g.single(); e == nil {
		t.FailNow()
	}
}

func TestGenerate(t *testing.T) {
	t.Parallel()
	g := Generator{Logger: &mockLogger{}}

	// reset defaults for parameters
	mockParseTemplate = template.New("test")
	mockCreateError = nil
	mockReadfileError = nil
	mockMkdirError = nil
	statError = nil
	notExist = false
	mockTemplateError = nil

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
	mockWalkError = mockError
	if e := g.Generate(); e == nil {
		t.FailNow()
	}

	// test template error
	mockParseError = mockError
	if e := g.Generate(); e == nil {
		t.FailNow()
	}
}
