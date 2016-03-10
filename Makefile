
build:
	go build ./cmd/fconf

start:build
	./fconf

dev:build
	./fconf -dev true
