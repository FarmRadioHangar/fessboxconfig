package config

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"unicode"
)

type TokenType int

var eof = rune(-1)

const (
	EOF TokenType = iota
	Comment
	Section
	WHiteSpace
	NewLine
	Ident
	Operand
	OpenBrace
	ClosingBrace
)

// Token is the identifier for a chuck of text.
type Token struct {
	Type   TokenType
	Text   string
	Line   int
	Column int
}

// Scanner is a lexical scanner for scaning configuration files.
type Scanner struct {
	r      *bufio.Reader
	txt    *bytes.Buffer
	line   int
	err    error
	column int
}

func NewScanner(src io.Reader) *Scanner {
	return &Scanner{
		r:   bufio.NewReader(src),
		txt: &bytes.Buffer{},
	}
}

func (s *Scanner) Scan() (*Token, error) {
	ch := s.peek()
	if isIdent(ch) {
		return s.scanIdent()
	}
	switch ch {
	case ';':
		return s.scanComment()
	case ' ', '\t':
		return s.scanWhitespace()
	case '\n', '\r':
		return s.scanNewline()
	case '=':
		return s.scanRune(Operand)
	case '[':
		return s.scanRune(OpenBrace)
	case ']':
		return s.scanRune(ClosingBrace)
	case eof:
		return nil, io.EOF
	}
	return nil, errors.New("unrecognized token " + string(ch))
}

func (s *Scanner) scanComment() (*Token, error) {
	tok := &Token{}
	buf := &bytes.Buffer{}
END:
	for {
		ch, size, err := s.r.ReadRune()
		if err != nil {
			if err.Error() == io.EOF.Error() {

				break END
			}
			return nil, err
		}
		switch ch {
		case '\n', '\r':
			s.r.UnreadRune()
			break END
		default:
			buf.WriteRune(ch)
			s.column = s.column + size
		}
	}
	tok.Column = s.column
	tok.Type = Comment
	tok.Text = buf.String()
	tok.Line = s.line
	return tok, nil
}
func (s *Scanner) scanWhitespace() (*Token, error) {
	tok := &Token{}
	buf := &bytes.Buffer{}
END:
	for {
		ch, _, err := s.r.ReadRune()
		if err != nil {
			if err.Error() == io.EOF.Error() {
				break END
			}
			return nil, err
		}
		switch ch {
		case ' ', '\t':
			buf.WriteRune(ch)
		default:
			s.r.UnreadRune()
			break END
		}
	}
	tok.Column = s.column
	tok.Type = Comment
	tok.Text = buf.String()
	tok.Line = s.line
	return tok, nil
}

func (s *Scanner) scanNewline() (*Token, error) {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return nil, err
	}
	tok := &Token{}
	tok.Type = NewLine
	tok.Text = string(ch)
	s.column = 0
	s.line++
	tok.Column = s.column
	tok.Line = s.line
	return tok, nil
}

func isIdent(ch rune) bool {
	return ch == '_' || ch == '-' || ch == '+' || unicode.IsLetter(ch) || unicode.IsDigit(ch)
}

func (s *Scanner) scanIdent() (*Token, error) {
	return s.scanRune(Ident)
}

func (s *Scanner) scanRune(typ TokenType) (*Token, error) {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return nil, err
	}
	tok := &Token{}
	tok.Type = typ
	tok.Text = string(ch)
	s.column++
	tok.Column = s.column
	tok.Line = s.line
	return tok, nil
}

func (s *Scanner) peek() rune {
	ch, _, err := s.r.ReadRune()
	defer s.r.UnreadRune()
	if err != nil {
		if err.Error() == io.EOF.Error() {
			return eof
		}
		panic(err)
	}
	return ch
}
