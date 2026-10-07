package emulator

import (
	"errors"
	"fmt"
	"strconv"
)

// Ограничения на число аргументов команд.
const (
	maxExitArgs = 1
	maxSaveArgs = 1
)

// Ошибки проверки аргументов команд.
var (
	errTooManyArgs    = errors.New("too many arguments")
	errMissingOperand = errors.New("missing operand")
)

// defaultCommands возвращает таблицу встроенных команд оболочки.
func defaultCommands() map[string]handler {
	return map[string]handler{
		"ls":       stub("ls"),
		"cd":       stub("cd"),
		"exit":     cmdExit,
		"vfs-save": cmdVfsSave,
	}
}

// stub создаёт команду-заглушку, которая выводит своё имя и аргументы.
func stub(name string) handler {
	return func(s *Shell, args []string) error {
		fmt.Fprintf(s.out, "%s: args %q\n", name, args)
		return nil
	}
}

// cmdExit завершает работу оболочки с необязательным числовым кодом.
func cmdExit(_ *Shell, args []string) error {
	if len(args) > maxExitArgs {
		return errTooManyArgs
	}
	if len(args) == 0 {
		return exitRequest{code: 0}
	}
	code, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("%s: numeric argument required", args[0])
	}
	return exitRequest{code: code}
}

// cmdVfsSave сохраняет текущее состояние VFS на диск в виде директории.
func cmdVfsSave(s *Shell, args []string) error {
	if len(args) == 0 {
		return errMissingOperand
	}
	if len(args) > maxSaveArgs {
		return errTooManyArgs
	}
	if err := s.fs.Save(args[0]); err != nil {
		return fmt.Errorf("%s: %w", args[0], err)
	}
	fmt.Fprintf(s.out, "VFS saved to %s\n", args[0])
	return nil
}
