#!/bin/bash

SCRIPT_DIR=$PWD/scripts


function install_systemd {
    cp -f $SCRIPT_DIR/fconf.service /lib/systemd/system/fconf.service
    systemctl enable fconf || true
    systemctl daemon-reload || true
}



id fconf &>/dev/null
if [[ $? -ne 0 ]]; then
    useradd -r -K USERGROUPS_ENAB=yes -M fconf -s /bin/false -d /etc/fconf
fi


if [[ -f /etc/os-release ]]; then
    which systemctl &>/dev/null
    if [[ $? -eq 0 ]]; then
      echo "INSTALLING fconf systemd service"
	    install_systemd
	    systemctl restart fconf || echo "WARNING: systemd not running."
    else
      echo "Need to install systemd"
    fi
fi
