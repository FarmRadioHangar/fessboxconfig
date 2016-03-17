package ast

type Node interface {
	Begin() int
	End() int
	Text() string
}

type Context struct {
	Head        Node
	Templates   []Node
	Assignments []AsignStmt
	Objects     []Object
}

type Template Context

type AsignStmt struct {
	Left  []Node
	Equal Node // =
	Right []Node
}

type Object struct {
	Left   []Node
	Assign Node // =>
	Right  []Node
}

type File struct {
	Comments    []Node
	Contextx    []Context
	Assignments []AsignStmt
	Objects     []Object
	Templates   []Template
}
