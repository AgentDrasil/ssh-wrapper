#!/bin/sh

set -e

mkdir -p /etc/keys
chmod 0700 /etc/keys
chown root:root /etc/keys

install -o root -g root -m 0400 /run/secrets/key /etc/keys/key
install -o root -g root -m 0400 /run/secrets/config.yaml /etc/ssh.config.yaml

mkdir -p /var/log/ssh-wrapper
chown 1000:1000 /var/log/ssh-wrapper

tail -f /dev/null
