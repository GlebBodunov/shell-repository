package emulator

import (
	"errors"
	"strings"

	"shellemu/src/vfs"
)

// errRemoveCwd возвращается при попытке удалить текущую директорию
// или одну из её родительских директорий.
var errRemoveCwd = errors.New("cannot remove the current directory")

// cmdRm удаляет файлы и директории VFS в памяти.
// Опция -r удаляет директории рекурсивно, -f скрывает ошибки
// об отсутствующих файлах.
func cmdRm(s *Shell, args []string) error {
	opts, paths, err := parseFlags(args, "rRf")
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		return errMissingOperand
	}
	recursive := opts['r'] || opts['R']
	for _, p := range paths {
		err := s.remove(vfs.Resolve(s.cwd, p), recursive)
		if errors.Is(err, vfs.ErrNotExist) && opts['f'] {
			continue
		}
		if err != nil {
			s.printErr("rm", p, err)
		}
	}
	return nil
}

// remove удаляет узел VFS по абсолютному пути, запрещая удаление
// текущей директории и её родителей.
func (s *Shell) remove(p string, recursive bool) error {
	if p != "/" && (s.cwd == p || strings.HasPrefix(s.cwd, p+"/")) {
		return errRemoveCwd
	}
	return s.fs.Remove(p, recursive)
}
