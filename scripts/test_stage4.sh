#!/bin/sh
# Стартовый скрипт этапа 4 с многоуровневой VFS.
cd "$(dirname "$0")/.."
./run.sh --vfs examples/vfs/deep --script scripts/emu/stage4.emu
