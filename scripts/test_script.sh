#!/bin/sh
# Запуск эмулятора только со стартовым скриптом.
cd "$(dirname "$0")/.."
./run.sh --script scripts/emu/stage2.emu
