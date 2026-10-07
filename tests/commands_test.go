package tests

import (
	"bytes"
	"strings"
	"testing"

	"shellemu/src/emulator"
	"shellemu/src/vfs"
)

// newDeepShell создаёт оболочку с VFS из примера deep.
func newDeepShell(t *testing.T) (*emulator.Shell, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	fsys, err := vfs.Load(deepVFS)
	if err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	return emulator.New(&out, &errOut, fsys), &out, &errOut
}

// run выполняет команды и возвращает обычный вывод и вывод ошибок.
func run(t *testing.T, lines ...string) (string, string) {
	t.Helper()
	sh, out, errOut := newDeepShell(t)
	for _, line := range lines {
		sh.Execute(line)
	}
	return out.String(), errOut.String()
}

// TestLs проверяет вывод содержимого директорий и ошибки ls.
func TestLs(t *testing.T) {
	out, errOut := run(t, "ls", "ls -l etc", "ls missing", "ls -z")
	for _, want := range []string{"etc/\nhome/\nvar/\n", "config.cfg"} {
		if !strings.Contains(out, want) {
			t.Errorf("ls output %q does not contain %q", out, want)
		}
	}
	if !strings.Contains(errOut, "no such file") || !strings.Contains(errOut, "invalid option") {
		t.Errorf("unexpected ls errors %q", errOut)
	}
}

// TestCd проверяет смену директории и приглашение.
func TestCd(t *testing.T) {
	sh, _, errOut := newDeepShell(t)
	sh.Execute("cd home/user/docs")
	if !strings.HasSuffix(sh.Prompt(), ":~/home/user/docs$ ") {
		t.Errorf("prompt after cd = %q", sh.Prompt())
	}
	sh.Execute("cd ../..")
	sh.Execute("cd /etc/config.cfg")
	sh.Execute("cd")
	if !strings.HasSuffix(sh.Prompt(), ":~$ ") {
		t.Errorf("prompt after cd = %q", sh.Prompt())
	}
	if !strings.Contains(errOut.String(), "not a directory") {
		t.Errorf("unexpected cd errors %q", errOut.String())
	}
}

// TestRev проверяет переворот строк файла.
func TestRev(t *testing.T) {
	out, errOut := run(t, "rev etc/config.cfg", "rev home")
	if !strings.Contains(out, "rotalume=emantsoh\n0808=trop\n") {
		t.Errorf("unexpected rev output %q", out)
	}
	if !strings.Contains(errOut, "is a directory") {
		t.Errorf("unexpected rev errors %q", errOut)
	}
}

// TestWc проверяет подсчёт строк, слов и байтов.
func TestWc(t *testing.T) {
	out, _ := run(t, "wc etc/config.cfg", "wc -l etc/config.cfg var/log/app/app.log")
	for _, want := range []string{"2       2      28 etc/config.cfg", "4 total"} {
		if !strings.Contains(out, want) {
			t.Errorf("wc output %q does not contain %q", out, want)
		}
	}
}

// TestWhoami проверяет вывод имени пользователя.
func TestWhoami(t *testing.T) {
	out, errOut := run(t, "whoami", "whoami x")
	if strings.TrimSpace(out) == "" || !strings.Contains(errOut, "extra operand") {
		t.Errorf("whoami output %q, errors %q", out, errOut)
	}
}
