#!/bin/sh
# Загрузка VFS "deep", сохранение через vfs-save и сравнение с исходником.
cd "$(dirname "$0")/.."
rm -rf saved_vfs/deep
SAVE_DIR=saved_vfs/deep ./run.sh --vfs examples/vfs/deep --script scripts/emu/vfs_save.emu
diff -r examples/vfs/deep saved_vfs/deep && echo "OK: saved VFS matches the source"
