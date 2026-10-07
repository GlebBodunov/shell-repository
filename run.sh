#!/bin/sh
# Собирает эмулятор и запускает его. Все аргументы передаются эмулятору.
set -e
DIR="$(cd "$(dirname "$0")" && pwd)"
(cd "$DIR" && go build -o bin/emulator ./src)
exec "$DIR/bin/emulator" "$@"
