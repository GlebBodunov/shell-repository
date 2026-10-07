#!/bin/sh
# Запуск эмулятора со всеми параметрами командной строки.
cd "$(dirname "$0")/.."
./run.sh --vfs examples/vfs/minimal --script scripts/emu/stage2.emu
