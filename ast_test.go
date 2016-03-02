package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestParser(t *testing.T) {
	src, err := ioutil.ReadFile("fixture/modem.txt")
	if err != nil {
		t.Fatal(err)
	}
	p, err := newParser(bytes.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	ass, err := p.parse()
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range ass.sections {
		fmt.Println(v.name)
		for _, i := range v.values {
			fmt.Printf(">> %s  %s\n", i.key, i.value)
		}
	}

}
