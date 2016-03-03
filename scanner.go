package config

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"unicode"
)

// TokenType is the type of tokem that will be returned by the Scanner.
type TokenType int

var eof = rune(-1)

// The following are the token types that are recognized by the scanner
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
// This works only on UTF-& text.
type Scanner struct {
	r      *bufio.Reader
	txt    *bytes.Buffer
	line   int
	err    error
	column int
}

// NewScanner takes src and returns a new Scanner.
func NewScanner(src io.Reader) *Scanner {
	return &Scanner{
		r:   bufio.NewReader(src),
		txt: &bytes.Buffer{},
	}
}

//Scan returns a new token for every call by advancing on the consumend UTF-8
//encoded input text.
//
// Anything after ; is considered a comment. White space is preserved to gether
// with  new lines. New lines and spaces are intepreted differently.
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

//scanComment scans the input for Comments, only single line comments are
//supported.
//
// A comment is all the text that is after a comment identifier, This does not
// enforce the identifier, so it is up to the caller to decied where the comment
// starts, this will read all the text up to the end of the line and return it
// as a single comment token.
//
// TODO(gernest) accept thecomment identifier, or check whether the first
// rune is the supported token identifier.
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

//scanWhitespace scans all utf-8 white space characters until it hits a non
//whitespace character.
//
// Tabs ('\t') and space(' ') all represent white space.
func (s *Scanner) scanWhitespace() (*Token, error) {
	tok := &Token{}

	// There can be arbitrary spaces so we need to bugger them up.
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
			// Stop after hitting non whitespace character
			// Reseting the buffer is necessary so that the scanned character can be
			// accessed for the next call to Scan method.
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

//scanNewline returns a token of type NewLine. It is necessary to separate
//newlines from normal spaces because many configuration files formats make use
//of new lines.
//
// A new line can either be a carriage return( '\r') or a new line
// character('\n')
//
// TODO(gernest) accept a new line character as input.
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

//isIdent returns true if ch is a valid identifier
// valid identifiers are
//	underscore _
//	dash -
//	plus +
//	a unicode letter a-zA-Z
//	a unicode digit 0-9
func isIdent(ch rune) bool {
	return ch == '_' || ch == '-' || ch == '+' || unicode.IsLetter(ch) || unicode.IsDigit(ch)
}

//scanIdent returns the current character in the input source as an Ident Token
//
// TODO(gernest) Accept the character as input argument.
func (s *Scanner) scanIdent() (*Token, error) {
	return s.scanRune(Ident)
}

// scanRune scans the current rune and returns a token of type typ, whose Text
// is the scanned character
//
// Use this for single character tokens
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

// peek returns the next rune in the input buffer but does not advance the
// position of the current buffer.
//
// This is a safe way to peek at the next  rune character without actually
// reading it.
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
