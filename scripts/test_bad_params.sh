#!/bin/sh
# Обработка ошибок в параметрах командной строки.
cd "$(dirname "$0")/.."
echo "== неизвестный параметр =="
./run.sh --unknown
echo "exit code: $?"
echo "== лишний позиционный аргумент =="
./run.sh extra
echo "exit code: $?"
echo "== несуществующий стартовый скрипт =="
./run.sh --script no_such_script.emu
echo "exit code: $?"
