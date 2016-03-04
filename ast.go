package config

import (
	"encoding/json"
	"errors"
	"io"
)

// ast is an abstract syntax tree for a configuration object. The configuration
// format should be section based( or you can say namespacing).
type ast struct {
	sections []*nodeSection
}

//Section returns the section named name or an error if the section is not found
//in the ast
func (a *ast) Section(name string) (*nodeSection, error) {
	for _, v := range a.sections {
		if v.name == name {
			return v, nil
		}
	}
	return nil, errors.New("section not found")
}

//ToJSON marhalls *ast to a json string and writes the result to dst
func (a *ast) ToJSON(dst io.Writer) error {
	o := make(map[string]interface{})
	for _, v := range a.sections {
		sec := make(map[string]interface{})
		for _, value := range v.values {
			sec[value.key] = value.value
		}
		o[v.name] = sec
	}
	return json.NewEncoder(dst).Encode(o)
}

//nodeSection represent a section in the configuration object. Sections are name
//spaces that contains configurations definitions under them.
type nodeSection struct {
	name   string
	line   int
	values []*nodeIdent
}

//Get access the key definition and returns its value or an error if the key is
//not part of the section.
func (n *nodeSection) Get(key string) (string, error) {
	for _, v := range n.values {
		if v.key == key {
			return v.value, nil
		}
	}
	return "", errors.New("key not found")
}

//nodeIdent represents a configuration definition, which can be the key value
//definition.
type nodeIdent struct {
	key   string
	value string
	line  int
}

// parser is a parser for configuration files. It supports utf-8 encoded
// configuration files.
//
// Only modem configuration files are supported for the momment.
type parser struct {
	tokens  []*Token
	ast     *ast
	currPos int
}

//newParser returns a new parser that parses input from src. The returned parser
//supports gsm modem configuration format only.
func newParser(src io.Reader) (*parser, error) {
	s := NewScanner(src)
	var toks []*Token
	var err error
	var tok *Token
	for err == nil {
		tok, err = s.Scan()
		if err != nil {
			if err.Error() != io.EOF.Error() {
				return nil, err
			}

			// we are at the end of the input
			break
		}
		if tok != nil {
			switch tok.Type {
			case WhiteSpace, Comment:

				// Skip comments and whitespaces but preserve the newlines to aid in
				// parsing
				continue
			default:
				toks = append(toks, tok)
			}

		}
	}
	return &parser{tokens: toks, ast: &ast{}}, nil
}

func (p *parser) parse() (*ast, error) {
	var err error
	if err != nil {
		return nil, err
	}
	mainSec := &nodeSection{name: "main"}
END:
	for {
		tok := p.next()
		if tok.Type == EOF {
			break END
		}
		switch tok.Type {
		case OpenBrace:
			p.rewind()
			err = p.parseSection()
			if err != nil {
				break END
			}
		case Ident:
			p.rewind()
			err = p.parseIdent(mainSec)
			if err != nil {
				break END
			}

		}
	}
	if err != nil {
		return nil, err
	}
	p.ast.sections = append([]*nodeSection{mainSec}, p.ast.sections...)
	return p.ast, err
}

func (p *parser) next() *Token {
	if p.currPos >= len(p.tokens)-1 {
		return &Token{Type: EOF}
	}
	t := p.tokens[p.currPos]
	p.currPos++
	return t
}

func (p *parser) seek(at int) {
	p.currPos = at
}

func (p *parser) parseSection() (err error) {
	left := p.next()
	if left.Type != OpenBrace {
		return errors.New("bad token")
	}
	ns := &nodeSection{}
	completeName := false
END:
	for {
	BEGIN:
		tok := p.next()
		if tok.Type == EOF {
			p.rewind()
			break END
		}

		if !completeName {
			switch tok.Type {
			case Ident:
				ns.name = ns.name + tok.Text
				goto BEGIN
			case ClosingBrace:
				completeName = true
				goto BEGIN
			}
		}
		switch tok.Type {
		case NewLine:
			n1 := p.next()
			if n1.Type == NewLine {
				n2 := p.next()
				if n2.Type == NewLine {
					break END
				}
				p.rewind()
				goto BEGIN
			}
			goto BEGIN
		case Ident:
			p.rewind()
			err = p.parseIdent(ns)
			if err != nil {
				break END
			}
		default:
			break END
		}
	}
	if err == nil {
		p.ast.sections = append(p.ast.sections, ns)
	}
	return
}

func (p *parser) rewind() {
	p.currPos--
}

func (p *parser) parseIdent(sec *nodeSection) (err error) {
	n := &nodeIdent{}
	doneKey := false
END:
	for {
	BEGIN:
		tok := p.next()
		if tok.Type == EOF {
			p.rewind()
			break END
		}

		if !doneKey {
			switch tok.Type {
			case Ident:
				n.key = n.key + tok.Text
				goto BEGIN
			case Operand:
				doneKey = true
				goto BEGIN
			default:
				err = errors.New("some fish")
				break END
			}

		}
		switch tok.Type {
		case Ident:
			n.value = n.value + tok.Text
			goto BEGIN
		case NewLine:
			break END
		default:
			err = errors.New("some fish")
			break END
		}
	}
	if err == nil {
		sec.values = append(sec.values, n)
	}

	return
}
