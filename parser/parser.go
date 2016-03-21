//Package gsm provides a parser for  asterisk configuration files
//
// The abstract syntax tree is basic, and made to accomodate the urgent need to
// parse the dongle configuration file and also as a means to see if scanning is done
// well will the scanner.
//
//TODO(gernest) rewrite the AST and move it into a separate package.
package parser

import (
	"fmt"
	"io"

	"github.com/FarmRadioHangar/fessboxconfig/ast"
	"github.com/FarmRadioHangar/fessboxconfig/scanner"
)

type node struct {
	begin, end int
	txt        string
}

func (n *node) Begin() int {
	return n.begin
}
func (n *node) Text() string {
	return n.txt
}
func (n *node) End() int {
	return n.end
}

// Parser is a Parser for scanneruration files. It supports utf-8 encoded
// scanneruration files.
//
// Only modem scanneruration files are supported for the momment.
type Parser struct {
	tokens  []*ast.Token
	currPos int
	Ass     *ast.File
}

//NewParser returns a new Parser that parses input from src. The returned Parser
//supports gsm modem scanneruration format only.
func NewParser(src io.Reader) (*Parser, error) {
	s := scanner.NewScanner(src)
	var toks []*ast.Token
	var err error
	var tok *ast.Token
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
			if tok.Type == ast.WhiteSpace {
				continue
			}
			toks = append(toks, tok)
		}
	}
	return &Parser{tokens: toks, Ass: &ast.File{}}, nil
}

// Parse parses the scanned input and return its *Ast or arror if any.
func (p *Parser) Parse() (*ast.File, error) {
	var perr error
END:
	for {
		peek := p.peek()
		if peek.Type == ast.EOF {
			break END
		}
		switch peek.Type {
		case ast.LBrace:
			ctx, err := p.context()
			if err != nil {
				perr = err
				break END
			}
			if len(ctx.Templates) == 1 {
				t := ctx.Templates[0]
				if t.Text() == "!" {
					p.Ass.Templates = append(p.Ass.Templates, ast.Template(ctx))
					continue
				}
			}
			p.Ass.Contexts = append(p.Ass.Contexts, ctx)
		case ast.Comment:
			tok := p.next()
			n := &node{
				begin: tok.Begin,
				end:   tok.End,
				txt:   tok.Text,
			}
			p.Ass.Comments = append(p.Ass.Comments, n)
		case ast.Ident:
			a, err := p.parseStmt()
			if err != nil {
				perr = err
				break END
			}
			p.Ass.Assignments = append(p.Ass.Assignments, a)
		}
	}
	if perr != nil {
		return nil, perr
	}
	return p.Ass, nil
}

func (p *Parser) parseStmt() (ast.AsignStmt, error) {
	a := ast.AsignStmt{}
	var isLeft = true
END:
	for {
		tok := p.next()
		switch tok.Type {
		case ast.EOF:
			break END
		case ast.Ident:
			n, err := p.parseIdent()
			if err != nil {
				break END
			}
			if isLeft {
				a.Left = append(a.Left, n)
				continue
			}
			a.Right = append(a.Right, n)
		case ast.Assign:
			peek := p.peek()
			n := &node{}
			n.begin = tok.Begin
			n.txt += tok.Text
			n.end = tok.End
			if peek.Type == ast.Greater {
				n.txt += peek.Text
				n.end = peek.End

				// advance current position to cover the next token
				p.currPos++
			}
			a.Equal = n
			isLeft = false
		case ast.NLine:
			break END
		}
	}
	return a, nil
}

func (p *Parser) parseIdent() (ast.Node, error) {
	n := &node{}
	tok := p.next()
	if tok.Type == ast.EOF {
		return nil, io.EOF
	}
	n.begin = tok.Begin
END:
	for {
		tok = p.next()
		if tok.Type == ast.EOF {
			return nil, io.EOF
		}
		if tok.Type != ast.Ident {
			break END
		}
		n.txt = tok.Text
	}
	return n, nil
}
func (p *Parser) context() (ast.Context, error) {
	ctx := ast.Context{}
END:
	for {
		tok := p.peek()
		switch tok.Type {
		case ast.LBrace:
			n, err := p.contextHead()
			if err != nil {
				break END
			}
			ctx.Head = n
		case ast.LBracket:
			tpl, err := p.contextTemplates()
			if err != nil {
				break END
			}
			ctx.Templates = append(ctx.Templates, tpl...)
		case ast.Comment:
			c := &node{
				begin: tok.Begin,
				end:   tok.End,
				txt:   tok.Text,
			}
			ctx.Comments = append(ctx.Comments, c)
		case ast.Ident:
			n, err := p.parseStmt()
			if err != nil {
				break END
			}
			if n.Equal.Text() == "=>" {
				ctx.Objects = append(ctx.Objects, ast.Object(n))
				continue
			}
			ctx.Assignments = append(ctx.Assignments, n)
		case ast.EOF:
			break END
		}
	}
	return ctx, nil
}

func (p *Parser) contextHead() (ast.Node, error) {
	n := &node{}
	begin := p.next()
	name, err := p.parseIdent()
	if err != nil {
		return nil, err
	}
	end := p.next()
	if end.Type != ast.RBrace {
		return nil, fmt.Errorf(" %d: %depected [ got %s instead ", end.Line, end.Column, end.Text)
	}
	n.begin = begin.Begin
	n.txt += begin.Text
	n.txt += name.Text()
	n.txt += end.Text
	n.end = end.End
	return n, nil
}

func (p *Parser) contextTemplates() ([]ast.Node, error) {
	return nil, nil
}

func (p *Parser) next() *ast.Token {
	if p.currPos >= len(p.tokens)-1 {
		return &ast.Token{Type: ast.EOF}
	}
	t := p.tokens[p.currPos]
	p.currPos++
	return t
}

func (p *Parser) seek(at int) {
	p.currPos = at
}

func (p *Parser) rewind() {
	p.currPos--
}

// Returns the next tokken without updating the current position. This is one
// step look ahead.
func (p *Parser) peek() *ast.Token {
	if p.currPos >= len(p.tokens)-1 {
		return &ast.Token{Type: ast.EOF}
	}
	return p.tokens[p.currPos]
}
