package lexer

import (
	"bytes"

	"github.com/FarmRadioHangar/combo"
)

const (
	Assign combo.TokenType = iota
	LeftBrace
	RightBrace
	Plain
	DoubleQuote
	SingleQuote
	Hash
	Equal
	Space
	NewLine
	Tab
	Eof
	Comment
)

var specialChars = struct {
	Assign      rune
	LeftBrace   rune
	RightBrace  rune
	DoubleQuote rune
	SingleQuote rune
	Hash        rune
	Space       rune
	NewLine     rune
	Tab         rune
}{
	'=', '[', ']', '"', '\'', '#', ' ', '\n', '\t',
}

func newline() combo.Lexer {
	return combo.RuneLex('\n', NewLine)
}

func space() combo.Lexer {
	return combo.RuneLex(' ', Space)
}

func tab() combo.Lexer {
	return combo.RuneLex('\t', Tab)
}

func equal() combo.Lexer {
	return combo.RuneLex('=', Equal)
}

func hash() combo.Lexer {
	return combo.RuneLex('#', Hash)
}

func doubleQuote() combo.Lexer {
	return combo.RuneLex('"', DoubleQuote)
}

func singleQuote() combo.Lexer {
	return combo.StringLex("'", SingleQuote)
}

func leftBrace() combo.Lexer {
	return combo.RuneLex('[', LeftBrace)
}
func rightBrace() combo.Lexer {
	return combo.RuneLex(']', LeftBrace)
}

func isSpecial(r rune) bool {
	switch r {
	case specialChars.Assign,
		specialChars.DoubleQuote,
		specialChars.SingleQuote,
		specialChars.Hash,
		specialChars.LeftBrace,
		specialChars.NewLine,
		specialChars.Space,
		specialChars.Tab,
		specialChars.RightBrace:
		return true
	}
	return false
}

func plain() combo.Lexer {
	return func(b *bytes.Reader) (combo.Token, error) {
		buf := &bytes.Buffer{}
		left := int(b.Size()) - b.Len()
		var right int
	END:
		for {
			ch, size, err := b.ReadRune()
			if err != nil {
				return nil, err
			}
			if isSpecial(ch) {
				b.UnreadRune()
				break END
			}
			right = right + size
			buf.WriteRune(ch)
		}
		return combo.NewToken(Plain, buf.String(), left, right), nil
	}
}

func theEnd() combo.Token {
	return combo.NewToken(Eof, "", 0, 0)
}
