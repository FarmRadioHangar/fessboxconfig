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

type Object AsignStmt

type File struct {
	Comments    []Node
	Contexts    []Context
	Assignments []AsignStmt
	Objects     []Object
	Templates   []Template
}
