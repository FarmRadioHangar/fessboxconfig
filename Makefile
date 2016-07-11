
build:
	go build -o fconf

start:build
	./fconf

dev:build
	./fconf -dev true

dep:
	go build --ldflags '-extldflags "-static"' -o fconf
