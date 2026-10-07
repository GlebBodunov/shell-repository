package tests

import (
	"reflect"
	"strings"
	"testing"

	"shellemu/src/emulator"
)

// TestStripComments проверяет удаление комментариев в синтаксисе Go.
func TestStripComments(t *testing.T) {
	text := "// comment\nls /a//b // tail\n/* block\nstill block */ cd x\n\n" +
		"cd /* inline */ y\nexit"
	want := []string{"ls /a//b", "cd x", "cd  y", "exit"}
	if got := emulator.StripComments(text); !reflect.DeepEqual(got, want) {
		t.Errorf("StripComments = %q, want %q", got, want)
	}
}

// TestRunScript проверяет вывод команд скрипта и остановку на exit.
func TestRunScript(t *testing.T) {
	sh, out, _ := newShell()
	done, code, err := sh.RunScript(strings.NewReader("ls a\nexit 4\ncd b\n"))
	if err != nil || !done || code != 4 {
		t.Fatalf("RunScript = (%v, %d, %v)", done, code, err)
	}
	text := out.String()
	if !strings.Contains(text, "$ ls a\n") || strings.Contains(text, "cd b") {
		t.Errorf("unexpected script output %q", text)
	}
}
