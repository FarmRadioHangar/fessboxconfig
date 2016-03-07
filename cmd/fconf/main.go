package main

import (
	"fmt"
	"io"
)

type DeviceConfig interface {
	Name() string
	LoadJSON(io.Reader) error
	Save() error
	ToJSON(io.Writer) error
}

func main() {
	fmt.Println("hello box")
}
