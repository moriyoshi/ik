package ik

import (
	"net/http"
	"strings"
	"testing"
)

type myOpener string

func (myOpener) FileSystem() http.FileSystem  { return http.Dir(".") }
func (myOpener) BasePath() string             { return "" }
func (myOpener) NewOpener(path string) Opener { return myOpener("") }
func (opener myOpener) NewLineReader(filename string) (LineReader, error) {
	return NewDefaultLineReader(filename, strings.NewReader(string(opener))), nil
}

func TestParseConfig(t *testing.T) {
	const data = "<test>\n" +
		"attr_name1 attr_value1\n" +
		"attr_name2 attr_value2\n" +
		"</test>\n"
	config, err := ParseConfig(myOpener(data), "test.cfg")
	if err != nil {
		panic(err.Error())
	}
	if len(config.Root.Elems) != 1 {
		t.Fail()
	}
	if config.Root.Elems[0].Name != "test" {
		t.Fail()
	}
	if len(config.Root.Elems[0].Attrs) != 2 {
		t.Fail()
	}
	if config.Root.Elems[0].Attrs["attr_name1"] != "attr_value1" {
		t.Fail()
	}
	if config.Root.Elems[0].Attrs["attr_name2"] != "attr_value2" {
		t.Fail()
	}
}

// vim: sts=4 sw=4 ts=4 noet
