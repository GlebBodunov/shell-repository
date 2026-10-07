package emulator

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"strings"

	"shellemu/src/vfs"
)

// handler — функция, реализующая команду оболочки.
type handler func(s *Shell, args []string) error

// exitRequest возвращается командой exit и содержит код завершения.
type exitRequest struct {
	code int
}

// Error реализует интерфейс error.
func (e exitRequest) Error() string {
	return fmt.Sprintf("exit %d", e.code)
}

// Shell — эмулятор командной оболочки с собственным набором команд.
type Shell struct {
	out      io.Writer
	errOut   io.Writer
	user     string
	host     string
	lookup   func(string) string
	commands map[string]handler
	fs       *vfs.FS
	cwd      string
}

// New создаёт оболочку, которая работает с виртуальной файловой системой
// fsys и пишет обычный вывод в out, а ошибки в errOut. Если fsys равна nil,
// используется пустая VFS. Имя пользователя и компьютера берутся из ОС.
func New(out, errOut io.Writer, fsys *vfs.FS) *Shell {
	if fsys == nil {
		fsys = vfs.New()
	}
	return &Shell{
		out:      out,
		errOut:   errOut,
		user:     currentUser(),
		host:     currentHost(),
		lookup:   os.Getenv,
		commands: defaultCommands(),
		fs:       fsys,
		cwd:      "/",
	}
}

// Prompt возвращает приглашение к вводу в формате username@hostname:путь$.
// Корень VFS отображается как ~.
func (s *Shell) Prompt() string {
	return fmt.Sprintf("%s@%s:%s$ ", s.user, s.host, displayPath(s.cwd))
}

// displayPath возвращает путь VFS для приглашения, заменяя корень на ~.
func displayPath(p string) string {
	if p == "/" {
		return "~"
	}
	return "~" + p
}

// Execute выполняет одну строку команды. Возвращает true и код завершения,
// если была вызвана команда exit.
func (s *Shell) Execute(line string) (bool, int) {
	args, err := Parse(line, s.lookup)
	if err != nil {
		fmt.Fprintf(s.errOut, "parse error: %v\n", err)
		return false, 0
	}
	if len(args) == 0 {
		return false, 0
	}
	cmd, ok := s.commands[args[0]]
	if !ok {
		fmt.Fprintf(s.errOut, "%s: command not found\n", args[0])
		return false, 0
	}
	err = cmd(s, args[1:])
	var exit exitRequest
	if errors.As(err, &exit) {
		return true, exit.code
	}
	if err != nil {
		fmt.Fprintf(s.errOut, "%s: %v\n", args[0], err)
	}
	return false, 0
}

// Run запускает интерактивный цикл чтения и выполнения команд (REPL).
// Возвращает код завершения оболочки.
func (s *Shell) Run(in io.Reader) int {
	scanner := bufio.NewScanner(in)
	for {
		fmt.Fprint(s.out, s.Prompt())
		if !scanner.Scan() {
			fmt.Fprintln(s.out)
			return 0
		}
		if done, code := s.Execute(scanner.Text()); done {
			return code
		}
	}
}

// currentUser возвращает имя текущего пользователя ОС.
func currentUser() string {
	if u, err := user.Current(); err == nil && u.Username != "" {
		return u.Username
	}
	if name := os.Getenv("USER"); name != "" {
		return name
	}
	return "user"
}

// currentHost возвращает короткое имя компьютера без доменной части.
func currentHost() string {
	name, err := os.Hostname()
	if err != nil || name == "" {
		return "localhost"
	}
	short, _, _ := strings.Cut(name, ".")
	return short
}
