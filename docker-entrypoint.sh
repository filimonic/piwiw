#!/bin/sh
set -e

update-ca-certificates

exec su-exec piwiw:piwiw /usr/local/bin/piwiw "$@"
