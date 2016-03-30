package ast

type Nodes []Node

func (n Nodes) Less(i, j int) bool {
	return n[i].Begin() < n[j].Begin()
}

func (n Nodes) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (n Nodes) Len() int {
	return len(n)
}
