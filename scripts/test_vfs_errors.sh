#!/bin/sh
# Ошибки загрузки VFS: несуществующий путь и файл вместо директории.
cd "$(dirname "$0")/.."
echo "== несуществующая директория =="
echo exit | ./run.sh --vfs examples/vfs/no_such_dir
echo "exit code: $?"
echo "== файл вместо директории =="
echo exit | ./run.sh --vfs examples/vfs/minimal/hello.txt
echo "exit code: $?"
