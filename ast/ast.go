package ast

type Node interface {
	Begin() int
	End() int
	Text() string
}

type Context struct {
	Head []Node
	Body []Node
}

type AsignStmt struct {
	Left  []Node
	Equal Node
	Right []Node
}
