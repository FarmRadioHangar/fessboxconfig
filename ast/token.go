package ast

// TokenType is the type of tokem that will be returned by the Scanner.
type TokenType int

// The following are the token types that are recognized by the scanner
const (
	EOF TokenType = iota
	Comment
	Section
	WhiteSpace
	NLine
	Ident
	Assign   // =
	LBrace   // [
	RBrace   // ]
	LBracket // )
	RBracket // (
	Exclam   // !
)

// Token is the identifier for a chunk of text.
type Token struct {
	Type   TokenType
	Text   string
	Line   int
	Column int
	Begin  int
	End    int
}
