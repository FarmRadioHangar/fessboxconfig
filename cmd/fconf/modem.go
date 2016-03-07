package main

import (
	"bytes"
	"io"
	"os"

	"github.com/FarmRadioHangar/fessboxconfig/gsm"
)

type modemConfig struct {
	name       string
	ast        *gsm.Ast
	backupName string
	dir        string
	fileName   string
	cache      []byte
}

func (m *modemConfig) Name() string {
	return m.name
}

func (m *modemConfig) LoadJSON(src io.Reader) error {
	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, src)
	if err != nil {
		return err
	}
	return m.ast.LoadJSON(buf.Bytes())
}

func (m *modemConfig) ToJSON(dst io.Writer) error {
	return m.ast.ToJSON(dst)
}

func (m *modemConfig) Save() error {
	name := filepath.Join(m.dir, m.fileName)
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_TRUNC, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	gsm.PrintAst(f, m.ast)
	return nil
}
