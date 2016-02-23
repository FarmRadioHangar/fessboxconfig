package lexer

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestLex(t *testing.T) {
	src, err := ioutil.ReadFile("fixture/modem.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, err = Lex(bytes.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
}
