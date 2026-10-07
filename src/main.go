// Command emulator запускает эмулятор командной оболочки UNIX-подобной ОС.
package main

import (
	"os"

	"shellemu/src/emulator"
)

// main создаёт оболочку и запускает её в интерактивном режиме.
func main() {
	sh := emulator.New(os.Stdout, os.Stderr)
	os.Exit(sh.Run(os.Stdin))
}
