package emulator

import (
	"errors"
	"fmt"
	"strconv"
)

// maxExitArgs — максимальное число аргументов команды exit.
const maxExitArgs = 1

// errTooManyArgs возвращается, если команде передано слишком много аргументов.
var errTooManyArgs = errors.New("too many arguments")

// defaultCommands возвращает таблицу встроенных команд оболочки.
func defaultCommands() map[string]handler {
	return map[string]handler{
		"ls":   stub("ls"),
		"cd":   stub("cd"),
		"exit": cmdExit,
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
