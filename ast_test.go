package config

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestParser(t *testing.T) {
	src, err := ioutil.ReadFile("fixture/modem.conf")
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
	mainSample := []struct {
		key, value string
	}{
		{"interval", "15"},
		{"group", "0"},
		{"language", "en"},
	}
	main, err := ass.Section("main")
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range mainSample {
		value, err := main.Get(v.key)
		if err != nil {
			t.Fatal(err)
		}
		if value != v.value {
			t.Errorf("expected %s got %s", v.value, value)
		}
	}

}
