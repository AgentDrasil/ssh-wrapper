#!/bin/sh

set -e
 
install -o root -g root -m 0400 /run/secrets/key /etc/key
install -o root -g root -m 0400 /run/secrets/config.yaml /etc/config.yaml
 
mkdir -p /var/log/ssh-wrapper
chown 1000:1000 /var/log/ssh-wrapper
 
tail -f /dev/null
