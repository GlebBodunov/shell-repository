package emulator

import (
	"fmt"
	"io/fs"
	"strings"

	"shellemu/src/vfs"
)

// lsOptions — режимы вывода команды ls.
type lsOptions struct {
	long bool
	all  bool
}

// cmdLs выводит содержимое директорий VFS или сведения о файлах.
// Поддерживаются опции -l (подробный вывод) и -a (скрытые файлы).
func cmdLs(s *Shell, args []string) error {
	opts, paths, err := parseFlags(args, "la")
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		paths = []string{"."}
	}
	ls := lsOptions{long: opts['l'], all: opts['a']}
	for i, p := range paths {
		if i > 0 {
			fmt.Fprintln(s.out)
		}
		if err := s.listPath(p, ls, len(paths) > singleOperand); err != nil {
			s.printErr("ls", p, err)
		}
	}
	return nil
}

// listPath выводит один путь для команды ls. Если withHeader равен true,
// перед содержимым директории печатается её имя.
func (s *Shell) listPath(p string, opts lsOptions, withHeader bool) error {
	node, err := s.fs.Lookup(vfs.Resolve(s.cwd, p))
	if err != nil {
		return err
	}
	if !node.IsDir {
		fmt.Fprintln(s.out, formatEntry(node, p, opts.long))
		return nil
	}
	if withHeader {
		fmt.Fprintf(s.out, "%s:\n", p)
	}
	for _, child := range node.Children() {
		if opts.all || !strings.HasPrefix(child.Name, ".") {
			fmt.Fprintln(s.out, formatEntry(child, child.Name, opts.long))
		}
	}
	return nil
}

// formatEntry форматирует строку вывода ls для узла с отображаемым именем name.
func formatEntry(node *vfs.Node, name string, long bool) string {
	mode := node.Mode
	if node.IsDir {
		mode |= fs.ModeDir
		name += "/"
	}
	if !long {
		return name
	}
	return fmt.Sprintf("%s %8d %s", mode, len(node.Data), name)
}

// cmdCd меняет текущую директорию VFS. Без аргументов переходит в корень.
func cmdCd(s *Shell, args []string) error {
	if len(args) > singleOperand {
		return errTooManyArgs
	}
	target := ""
	if len(args) == singleOperand {
		target = args[0]
	}
	p := vfs.Resolve(s.cwd, target)
	node, err := s.fs.Lookup(p)
	if err != nil {
		return fmt.Errorf("%s: %w", target, err)
	}
	if !node.IsDir {
		return fmt.Errorf("%s: %w", target, vfs.ErrNotDir)
	}
	s.cwd = p
	return nil
}

// parseFlags отделяет опции вида -abc от операндов. allowed содержит
// допустимые буквы опций. Аргумент -- завершает список опций.
func parseFlags(args []string, allowed string) (map[rune]bool, []string, error) {
	opts := map[rune]bool{}
	for i, arg := range args {
		if arg == "--" {
			return opts, args[i+1:], nil
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			return opts, args[i:], nil
		}
		for _, r := range arg[1:] {
			if !strings.ContainsRune(allowed, r) {
				return nil, nil, fmt.Errorf("invalid option -- '%c'", r)
			}
			opts[r] = true
		}
	}
	return opts, nil, nil
}

// printErr выводит сообщение об ошибке команды cmd для операнда target.
func (s *Shell) printErr(cmd, target string, err error) {
	fmt.Fprintf(s.errOut, "%s: %s: %v\n", cmd, target, err)
}
