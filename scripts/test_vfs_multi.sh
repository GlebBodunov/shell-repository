#!/bin/sh
# Загрузка VFS "multi", сохранение через vfs-save и сравнение с исходником.
cd "$(dirname "$0")/.."
rm -rf saved_vfs/multi
SAVE_DIR=saved_vfs/multi ./run.sh --vfs examples/vfs/multi --script scripts/emu/vfs_save.emu
diff -r examples/vfs/multi saved_vfs/multi && echo "OK: saved VFS matches the source"
