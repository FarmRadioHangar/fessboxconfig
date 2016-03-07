package main

import "io"

type DeviceConfig interface {
	Name() string
	LoadJSON(io.Reader) error
	Save() error
	ToJSON(io.Writer) error
}
