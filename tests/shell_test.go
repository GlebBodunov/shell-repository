package tests

import (
	"bytes"
	"strings"
	"testing"

	"shellemu/src/emulator"
)

// newShell создаёт оболочку с буферами для вывода и ошибок.
func newShell() (*emulator.Shell, *bytes.Buffer, *bytes.Buffer) {
	var out, errOut bytes.Buffer
	return emulator.New(&out, &errOut, nil), &out, &errOut
}

// TestPrompt проверяет формат приглашения username@hostname:~$.
func TestPrompt(t *testing.T) {
	sh, _, _ := newShell()
	p := sh.Prompt()
	if !strings.Contains(p, "@") || !strings.HasSuffix(p, "$ ") {
		t.Errorf("unexpected prompt %q", p)
	}
}

// TestUnknownCommand проверяет сообщение о неизвестной команде.
func TestUnknownCommand(t *testing.T) {
	sh, _, errOut := newShell()
	sh.Execute("foo")
	if !strings.Contains(errOut.String(), "foo: command not found") {
		t.Errorf("unexpected error output %q", errOut.String())
	}
}

// TestExit проверяет завершение работы и обработку ошибок exit.
func TestExit(t *testing.T) {
	sh, _, errOut := newShell()
	if done, code := sh.Execute("exit 3"); !done || code != 3 {
		t.Errorf("exit 3 = (%v, %d)", done, code)
	}
	if done, _ := sh.Execute("exit abc"); done {
		t.Error("exit abc must not terminate the shell")
	}
	if done, _ := sh.Execute("exit 1 2"); done {
		t.Error("exit 1 2 must not terminate the shell")
	}
	if !strings.Contains(errOut.String(), "numeric argument required") {
		t.Errorf("unexpected error output %q", errOut.String())
	}
}

// TestRun проверяет работу REPL до команды exit.
func TestRun(t *testing.T) {
	sh, out, _ := newShell()
	code := sh.Run(strings.NewReader("whoami\nexit 5\nls\n"))
	if code != 5 {
		t.Errorf("Run returned %d, want 5", code)
	}
	if strings.Count(out.String(), "$ ") != 2 {
		t.Error("commands after exit must not run")
	}
}

// TestVfsSaveErrors проверяет обработку ошибок команды vfs-save.
func TestVfsSaveErrors(t *testing.T) {
	sh, _, errOut := newShell()
	sh.Execute("vfs-save")
	sh.Execute("vfs-save a b")
	text := errOut.String()
	if !strings.Contains(text, "missing operand") || !strings.Contains(text, "too many") {
		t.Errorf("unexpected error output %q", text)
	}
}
