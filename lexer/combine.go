package lexer

import (
	"bufio"
	"bytes"
	"io"

	"github.com/FarmRadioHangar/combo"
)

func Lex(src io.Reader) ([]combo.Token, error) {
	c := NewCombinator()
	return lineLex(src, c)
}

func lineLex(src io.Reader, c *combo.LexCombinator) ([]combo.Token, error) {
	var result []combo.Token
	scan := bufio.NewReader(src)
	var serr error
	for {
		line, err := scan.ReadBytes('\n')
		if err != nil {
			serr = err
			break
		}
		toks, err := c.Lex(line)
		if err != nil {
			serr = err
			break
		}
		result = append(result, toks...)
	}
	if serr != nil {
		if serr.Error() == io.EOF.Error() {
			if result != nil {
				return result, nil
			}
		}
		return nil, serr
	}
	return result, nil
}

func comment(b *bytes.Reader) (combo.Token, error) {
	left := int(b.Size()) - b.Len()
	ch, size, err := b.ReadRune()
	if err != nil {
		b.UnreadRune()
		return nil, err
	}
	if ch == '#' {
		r := bufio.NewReader(b)
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		right := left + len(line) + size
		return combo.NewToken(Comment, string(ch)+line, left, right), nil
	}
	return nil, combo.ErrorMSG("bad token", left, string(ch), "#")
}

func nameSpace(c *combo.LexCombinator) combo.Lexer {
	return c.ChainAnd(leftBrace(), c.ChainOr(whiteSpace(c)), plain(),
		c.ChainOr(whiteSpace(c)), rightBrace(), whiteSpace(c))
}

func whiteSpace(c *combo.LexCombinator) combo.Lexer {
	return c.ChainOr(space(), tab(), newline())
}

func statement(c *combo.LexCombinator) combo.Lexer {
	return c.ChainAnd(plain(), equal(), plain())
}

func NewCombinator() *combo.LexCombinator {
	c := &combo.LexCombinator{}
	c.ChainOr(comment, whiteSpace(c), statement(c))
	return c
}
