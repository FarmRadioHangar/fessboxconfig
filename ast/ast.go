package ast

//Node is an interface represent a meaningful chunch of text in an input stream
type Node interface {
	Begin() int
	End() int
	Text() string
}

//Context is a scetion in asterisk configuration file. It holds all the
//information of the section.
type Context struct {
	Head        Node
	Templates   []Node
	Assignments []AsignStmt
	Objects     []Object
	Comments    []Node
}

//Template is a Context but that is used as a template to composing other
//contexts by the asterisk parser.
type Template Context

//AsignStmt holds information about assignmant definitions in the configuration
//file.
type AsignStmt struct {
	Left  []Node
	Equal Node // =
	Right []Node
}

//Object is a special kind of assignment statement
type Object AsignStmt

//File is the abstract representation of the configuration file
type File struct {
	Comments    []Node
	Contexts    []Context
	Assignments []AsignStmt
	Objects     []Object
	Templates   []Template
}
