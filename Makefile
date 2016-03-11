
build:
	go build -o fconf

start:build
	./fconf

dev:build
	./fconf -dev true
