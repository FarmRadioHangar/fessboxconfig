CONFIG_DIR=/etc/fconf
BIN_DIR=/usr/bin
build:
	go build -o fconf

start:build
	./fconf

dev:build
	./fconf -dev true

prepare:fconf
	mkdir -p $(CONFIG_DIR)
	cp -i ./etc/fconf.json $(CONFIG_DIR)/fconf.json
	cp -i ./fconf $(BIN_DIR)/fconf

uninstall:
	which systemctl||true
	rm -r $(CONFIG_DIR)
	rm -f $(BIN_DIR)/fconf
	systemctl disable fconf
	rm -f /lib/systemd/system/fconf.service

install:prepare
	./scripts/install.sh
