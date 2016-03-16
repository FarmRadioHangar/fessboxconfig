package parser

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestParser(t *testing.T) {
	src, err := ioutil.ReadFile("modem.conf")
	if err != nil {
		t.Fatal(err)
	}
	p, err := NewParser(bytes.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	ass, err := p.Parse()
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

	// test the encoding to json
	dst := &bytes.Buffer{}
	err = ass.ToJSON(dst)
	if err != nil {
		t.Error(err)
	}
	nAst := &Ast{}
	err = nAst.LoadJSON(dst.Bytes())
	if err != nil {
		t.Error(err)
	}

	main, err = nAst.Section("main")
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
	nBuf := &bytes.Buffer{}
	PrintAst(nBuf, nAst)

	np, err := NewParser(nBuf)
	if err != nil {
		t.Fatal(err)
	}
	_, err = np.Parse()
	if err != nil {
		t.Error(err)
	}
}
