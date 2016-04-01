package main

import "github.com/jochenvg/go-udev"

type device struct {
	name string
	dev  *udev.Device
}

type manager struct {
	devices map[string]*device
	stop    chan struct{}
	m       *udev.Monitor
	u       *udev.Udev
}

func (mg *manager) Close() {
	mg.stop <- struct{}{}
}

func (mg *manager) Watch() {
}
