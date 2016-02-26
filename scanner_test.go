package config

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"testing"
)

func TestScanner(t *testing.T) {
	src, err := ioutil.ReadFile("fixture/modem.txt")
	if err != nil {
		t.Fatal(err)
	}
	s := NewScanner(bytes.NewReader(src))
	var tok *Token
	for err == nil {
		tok, err = s.Scan()
		if err != nil {
			if err.Error() != io.EOF.Error() {
				t.Fatal(err)
			}
		}
		if tok != nil {
			fmt.Println(tok.Line, ":", tok.Column, " ", tok.Text)
		}
	}
}
