package staticmd

import (
	"errors"
	"os"
	"testing"
)

func init() {
	runnable = mockCmd{}
	stat = func(_ string) (os.FileInfo, error) { return nil, statError }
	isNotExist = func(_ error) bool { return notExist }
}

var mockError = errors.New("mock error")

var mockCmdByteArray []byte
var mockCmdError error
var statError error
var notExist bool

type mockCmd struct{}

func (self mockCmd) Run(command string, args ...string) ([]byte, error) {
	return mockCmdByteArray, mockCmdError
}

func TestPlacebo(_ *testing.T) {}

func TestCmd(t *testing.T) {
	t.Parallel()
	c := cmd{}
	if _, e := c.Run(""); e == nil {
		t.FailNow()
	}
}

func TestExists(t *testing.T) {
	t.Parallel()

	// test stat success exists fail
	if _, e := exists(""); e != nil {
		t.FailNow()
	}

	// test stat success exists success
	notExist = true
	if _, e := exists(""); e != nil {
		t.FailNow()
	}

	// test stat fail
	statError = mockError
	if _, e := exists(""); e != nil {
		t.FailNow()
	}

	// test stat fail exists success
	notExist = false
	if _, e := exists(""); e == nil {
		t.FailNow()
	}
}

func TestVersion(t *testing.T) {
	t.Parallel()

	// test with byte array
	compare := "newp"
	mockCmdByteArray = []byte(compare)
	if v := version(""); v != compare {
		t.FailNow()
	}

	// test with error
	mockCmdError = mockError
	if v := version(""); v == compare || len(v) == 0 {
		t.FailNow()
	}
}

func TestBasename(t *testing.T) {
	t.Parallel()

	f := "/some/long/path"
	if o := basename(f); len(o) == 0 {
		t.FailNow()
	}
}

func TestIsMarkdown(t *testing.T) {
	t.Parallel()

	// test matching types
	for i := range extensions {
		if !isMarkdown("file." + extensions[i]) {
			t.FailNow()
		}
	}

	// test non matching type
	if isMarkdown("nope") {
		t.FailNow()
	}
}
