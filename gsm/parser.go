//Package gsm provides a parser for dongle configuration file for asterisk. The
//configuration format is a subset of the astersk dial plan.
//
// The abstract syntax tree is basic, and made to accomodate the urgent need to
// parse the configuration file and also as a means to see if scanning is done
// well will the scanner.
//
//TODO(gernest) rewrite the AST and move it into a separate package.
package gsm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	config "github.com/FarmRadioHangar/fessboxconfig/scanner"
)

// Ast is an abstract syntax tree for a configuration object. The configuration
// format should be section based( or you can say namespacing).
type Ast struct {
	sections []*NodeSection
}

//Section returns the section named name or an error if the section is not found
//in the Ast
func (a *Ast) Section(name string) (*NodeSection, error) {
	for _, v := range a.sections {
		if v.name == name {
			return v, nil
		}
	}
	return nil, errors.New("section not found")
}

//ToJSON marhalls *Ast to a json string and writes the result to dst
func (a *Ast) ToJSON(dst io.Writer) error {
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

//LoadJSON loads AST from json src
func (a *Ast) LoadJSON(src []byte) error {
	var obj map[string]interface{}
	err := json.Unmarshal(src, &obj)
	if err != nil {
		return err
	}
	for key, value := range obj {
		ns := &NodeSection{name: key}
		switch value.(type) {
		case map[string]interface{}:
			sec := value.(map[string]interface{})
			for k, v := range sec {
				ident := &nodeIdent{}
				ident.key = k
				ident.value = fmt.Sprint(v)
				ns.values = append(ns.values, ident)
			}
		}
		a.sections = append(a.sections, ns)
	}
	return nil
}

func PrintAst(dst io.Writer, src *Ast) {
	for _, v := range src.sections {
		if v.name == "main" {
			fmt.Fprintf(dst, "\n\n")
			for _, sub := range v.values {
				fmt.Fprintf(dst, "%s=%s \n", sub.key, sub.value)
			}
			fmt.Fprint(dst, "\n\n")
			continue
		}
		fmt.Fprintf(dst, "\n [%s]\n", v.name)
		for _, sub := range v.values {
			fmt.Fprintf(dst, "%s=%s \n", sub.key, sub.value)
		}
		fmt.Fprint(dst, "\n\n")
		continue
	}
}

//NodeSection represent a section in the configuration object. Sections are name
//spaces that contains configurations definitions under them.
type NodeSection struct {
	name   string
	line   int
	values []*nodeIdent
}

//Get access the key definition and returns its value or an error if the key is
//not part of the section.
func (n *NodeSection) Get(key string) (string, error) {
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

// Parser is a Parser for configuration files. It supports utf-8 encoded
// configuration files.
//
// Only modem configuration files are supported for the momment.
type Parser struct {
	tokens  []*config.Token
	Ast     *Ast
	currPos int
}

//NewParser returns a new Parser that parses input from src. The returned Parser
//supports gsm modem configuration format only.
func NewParser(src io.Reader) (*Parser, error) {
	s := config.NewScanner(src)
	var toks []*config.Token
	var err error
	var tok *config.Token
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
			case config.WhiteSpace, config.Comment:

				// Skip comments and whitespaces but preserve the newlines to aid in
				// parsing
				continue
			default:
				toks = append(toks, tok)
			}

		}
	}
	return &Parser{tokens: toks, Ast: &Ast{}}, nil
}

// Parse parses the scanned input and return its *Ast or arror if any.
func (p *Parser) Parse() (*Ast, error) {
	var err error
	if err != nil {
		return nil, err
	}
	mainSec := &NodeSection{name: "main"}
END:
	for {
		tok := p.next()
		if tok.Type == config.EOF {
			break END
		}
		switch tok.Type {
		case config.OpenBrace:
			p.rewind()
			err = p.parseSection()
			if err != nil {
				break END
			}
		case config.Ident:
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
	p.Ast.sections = append([]*NodeSection{mainSec}, p.Ast.sections...)
	return p.Ast, err
}

func (p *Parser) next() *config.Token {
	if p.currPos >= len(p.tokens)-1 {
		return &config.Token{Type: config.EOF}
	}
	t := p.tokens[p.currPos]
	p.currPos++
	return t
}

func (p *Parser) seek(at int) {
	p.currPos = at
}

func (p *Parser) parseSection() (err error) {
	left := p.next()
	if left.Type != config.OpenBrace {
		return errors.New("bad token")
	}
	ns := &NodeSection{}
	completeName := false
END:
	for {
	BEGIN:
		tok := p.next()
		if tok.Type == config.EOF {
			p.rewind()
			break END
		}

		if !completeName {
			switch tok.Type {
			case config.Ident:
				ns.name = ns.name + tok.Text
				goto BEGIN
			case config.ClosingBrace:
				completeName = true
				goto BEGIN
			}
		}
		switch tok.Type {
		case config.NewLine:
			n1 := p.next()
			if n1.Type == config.NewLine {
				n2 := p.next()
				if n2.Type == config.NewLine {
					break END
				}
				p.rewind()
				goto BEGIN
			}
			goto BEGIN
		case config.Ident:
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
		p.Ast.sections = append(p.Ast.sections, ns)
	}
	return
}

func (p *Parser) rewind() {
	p.currPos--
}

func (p *Parser) parseIdent(sec *NodeSection) (err error) {
	n := &nodeIdent{}
	doneKey := false
END:
	for {
	BEGIN:
		tok := p.next()
		if tok.Type == config.EOF {
			p.rewind()
			break END
		}

		if !doneKey {
			switch tok.Type {
			case config.Ident:
				n.key = n.key + tok.Text
				goto BEGIN
			case config.Operand:
				doneKey = true
				goto BEGIN
			default:
				err = errors.New("some fish")
				break END
			}

		}
		switch tok.Type {
		case config.Ident:
			n.value = n.value + tok.Text
			goto BEGIN
		case config.NewLine:
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
