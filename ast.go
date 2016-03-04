package config

import (
	"errors"
	"fmt"
	"io"
)

// ast is an abstract syntax tree for a configuration object. The configuration
// format should be section based( or you can say namespacing).
type ast struct {
	sections []*nodeSection
}

type nodeSection struct {
	name   string
	line   int
	values []*nodeIdent
}

type nodeIdent struct {
	key   string
	value string
	line  int
}

type parser struct {
	tokens  []*Token
	ast     *ast
	currPos int
}

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
			break
		}
		if tok != nil {
			switch tok.Type {
			case WhiteSpace, Comment:
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
		fmt.Println("parsing")
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
	fmt.Printf("parsing ident for %s -- ", sec.name)
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
	fmt.Println("done")

	return
}
