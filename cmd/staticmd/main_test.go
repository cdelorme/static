package main

import (
	"os"
	"testing"

	"github.com/cdelorme/staticmd"
)

func init() {
	exit = func(_ int) {}
	getwd = func() (string, error) { return "", nil }
}

var mockError error

type mockGenerator struct{}

func (self *mockGenerator) Generate() error { return mockError }

func TestPlacebo(_ *testing.T) {}

func TestMain(_ *testing.T) {
	os.Args = []string{}
	main()
}

func TestConfigure(t *testing.T) {

	// set a value on all parameters
	os.Args = []string{"-t", "afile", "-i", "/in/", "-o", "/out/", "-b", "-r"}

	// run configure & check results
	s := configure()
	if s == nil {
		t.FailNow()
	}

	// cast and check values on s
	g, e := s.(*staticmd.Generator)
	if !e {
		t.FailNow()
	}

	// check values on generator match cli parameters
	if g.Input != "/in/" || g.Output != "/out/" || !g.Relative || !g.Book || g.TemplateFile != "afile" {
		t.FailNow()
	}
}
