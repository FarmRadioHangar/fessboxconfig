package lexer

import "testing"

func TestIsSpecial(t *testing.T) {
	ok := isSpecial('#')
	if !ok {
		t.Error("epxpected true")
	}
}
