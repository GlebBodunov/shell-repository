#!/bin/sh
# Загрузка VFS "minimal", сохранение через vfs-save и сравнение с исходником.
cd "$(dirname "$0")/.."
rm -rf saved_vfs/minimal
SAVE_DIR=saved_vfs/minimal ./run.sh --vfs examples/vfs/minimal --script scripts/emu/vfs_save.emu
diff -r examples/vfs/minimal saved_vfs/minimal && echo "OK: saved VFS matches the source"
