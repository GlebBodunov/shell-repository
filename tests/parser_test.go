package tests

import (
	"errors"
	"reflect"
	"testing"

	"shellemu/src/emulator"
)

// fakeEnv возвращает значения тестовых переменных окружения.
func fakeEnv(name string) string {
	values := map[string]string{"HOME": "/home/test", "USER": "test"}
	return values[name]
}

// TestParse проверяет разбиение строки на слова и раскрытие переменных.
func TestParse(t *testing.T) {
	cases := []struct {
		line string
		want []string
	}{
		{"ls -l /tmp", []string{"ls", "-l", "/tmp"}},
		{"  cd   dir  ", []string{"cd", "dir"}},
		{"echo $HOME", []string{"echo", "/home/test"}},
		{"echo ${USER}x", []string{"echo", "testx"}},
		{"echo \"$HOME/docs\"", []string{"echo", "/home/test/docs"}},
		{"echo '$HOME'", []string{"echo", "$HOME"}},
		{"echo \"a b\" c", []string{"echo", "a b", "c"}},
		{"echo a\\ b", []string{"echo", "a b"}},
		{"echo $UNSET", []string{"echo"}},
		{"echo \"\"", []string{"echo", ""}},
		{"echo $", []string{"echo", "$"}},
		{"", nil},
	}
	for _, c := range cases {
		got, err := emulator.Parse(c.line, fakeEnv)
		if err != nil {
			t.Fatalf("Parse(%q): unexpected error %v", c.line, err)
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("Parse(%q) = %q, want %q", c.line, got, c.want)
		}
	}
}

// TestParseErrors проверяет обработку синтаксических ошибок.
func TestParseErrors(t *testing.T) {
	cases := []struct {
		line string
		want error
	}{
		{"echo 'abc", emulator.ErrUnterminatedQuote},
		{"echo \"abc", emulator.ErrUnterminatedQuote},
		{"echo ${HOME", emulator.ErrBadSubstitution},
		{"echo ${}", emulator.ErrBadSubstitution},
	}
	for _, c := range cases {
		if _, err := emulator.Parse(c.line, fakeEnv); !errors.Is(err, c.want) {
			t.Errorf("Parse(%q) error = %v, want %v", c.line, err, c.want)
		}
	}
}
