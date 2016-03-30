package parser

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestParser(t *testing.T) {
	src, err := ioutil.ReadFile("modem.conf")
	if err != nil {
		t.Error(err)
	}
	p, err := NewParser(bytes.NewReader(src))
	if err != nil {
		t.Error(err)
	}
	_, err = p.Parse()
	if err != nil {
		t.Error(err)
	}
}
