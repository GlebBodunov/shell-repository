#!/bin/sh
# Стартовый скрипт этапа 5. Исходная VFS на диске не изменяется.
cd "$(dirname "$0")/.."
rm -rf saved_vfs/stage5
./run.sh --vfs examples/vfs/deep --script scripts/emu/stage5.emu
echo "== исходная VFS на диске не изменилась =="
find examples/vfs/deep -type f | sort
echo "== сохранённое состояние после rm =="
find saved_vfs/stage5 | sort
