CONFIG_DIR=/etc/fconf
BIN_DIR=/usr/bin
build:
	go build -o fconf

start:build
	./fconf

dev:build
	./fconf -dev true

prepare:build
	mkdir -p $(CONFIG_DIR)
	cp -i -u ./etc/fconf.json $(CONFIG_DIR)/fconf.json
	cp -i -u ./fconf $(BIN_DIR)/fconf

uninstall:
	which systemctl||true
	rm -r $(CONFIG_DIR)
	rm -f $(BN_DIR)/fconf
	systemctl disable fconf
	rm -f /lib/systemd/system/fconf.service

install:prepare
	./scripts/install.sh
