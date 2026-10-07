#!/bin/sh
# Стартовый скрипт этапа 3 с многоуровневой VFS.
cd "$(dirname "$0")/.."
rm -rf saved_vfs/stage3
./run.sh --vfs examples/vfs/deep --script scripts/emu/stage3.emu
