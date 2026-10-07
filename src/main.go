// Command emulator запускает эмулятор командной оболочки UNIX-подобной ОС.
//
// Параметры командной строки:
//
//	--vfs PATH     путь к физическому расположению VFS
//	--script PATH  путь к стартовому скрипту
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"shellemu/src/emulator"
	"shellemu/src/vfs"
)

// Коды завершения программы при ошибках запуска.
const (
	exitConfigError = 2
	exitScriptError = 1
	exitVFSError    = 1
)

// config содержит параметры запуска эмулятора.
type config struct {
	vfsPath    string
	scriptPath string
}

// main разбирает параметры, выполняет стартовый скрипт и запускает REPL.
func main() {
	cfg, err := parseConfig(os.Args[1:], os.Stderr)
	if err != nil {
		os.Exit(exitConfigError)
	}
	printConfig(os.Stdout, cfg)
	os.Exit(run(cfg))
}

// parseConfig разбирает параметры командной строки.
func parseConfig(args []string, errOut io.Writer) (config, error) {
	var cfg config
	fs := flag.NewFlagSet("emulator", flag.ContinueOnError)
	fs.SetOutput(errOut)
	fs.StringVar(&cfg.vfsPath, "vfs", "", "путь к физическому расположению VFS")
	fs.StringVar(&cfg.scriptPath, "script", "", "путь к стартовому скрипту")
	if err := fs.Parse(args); err != nil {
		return cfg, err
	}
	if fs.NArg() > 0 {
		err := fmt.Errorf("unexpected arguments: %q", fs.Args())
		fmt.Fprintln(errOut, err)
		fs.Usage()
		return cfg, err
	}
	return cfg, nil
}

// printConfig выводит отладочную информацию о заданных параметрах.
func printConfig(w io.Writer, cfg config) {
	fmt.Fprintln(w, "[debug] emulator parameters:")
	fmt.Fprintf(w, "[debug]   vfs    = %s\n", valueOrUnset(cfg.vfsPath))
	fmt.Fprintf(w, "[debug]   script = %s\n", valueOrUnset(cfg.scriptPath))
}

// valueOrUnset возвращает значение параметра или пометку о его отсутствии.
func valueOrUnset(value string) string {
	if value == "" {
		return "(not set)"
	}
	return value
}

// run загружает VFS, выполняет стартовый скрипт, если он задан,
// и запускает REPL. Возвращает код завершения программы.
func run(cfg config) int {
	fsys, err := loadVFS(cfg.vfsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "vfs error: %s: %v\n", cfg.vfsPath, err)
		return exitVFSError
	}
	sh := emulator.New(os.Stdout, os.Stderr, fsys)
	if cfg.scriptPath != "" {
		done, code, err := runScript(sh, cfg.scriptPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "script error: %v\n", err)
			return exitScriptError
		}
		if done {
			return code
		}
	}
	return sh.Run(os.Stdin)
}

// runScript открывает файл стартового скрипта и выполняет его.
func runScript(sh *emulator.Shell, path string) (bool, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, 0, err
	}
	defer f.Close()
	return sh.RunScript(f)
}

// loadVFS загружает VFS из директории path. Если путь не задан,
// создаётся пустая VFS.
func loadVFS(path string) (*vfs.FS, error) {
	if path == "" {
		return vfs.New(), nil
	}
	fsys, err := vfs.Load(path)
	if err != nil {
		return nil, err
	}
	dirs, files := fsys.Stats()
	fmt.Printf("[debug] VFS loaded: %d directories, %d files\n", dirs, files)
	return fsys, nil
}
