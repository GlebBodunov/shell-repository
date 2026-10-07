package emulator

import (
	"fmt"
	"io"
	"strings"
)

// markerLen — длина маркеров комментариев //, /* и */.
const markerLen = 2

// RunScript выполняет команды из стартового скрипта. Каждая команда
// выводится вместе с приглашением, имитируя диалог с пользователем.
// Комментарии записываются в синтаксисе Go: // и /* ... */.
// Возвращает true и код завершения, если скрипт вызвал exit.
func (s *Shell) RunScript(r io.Reader) (bool, int, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return false, 0, err
	}
	for _, line := range StripComments(string(data)) {
		fmt.Fprintf(s.out, "%s%s\n", s.Prompt(), line)
		if done, code := s.Execute(line); done {
			return true, code, nil
		}
	}
	return false, 0, nil
}

// StripComments удаляет из текста скрипта комментарии в стиле Go
// и возвращает непустые строки с командами.
func StripComments(text string) []string {
	var lines []string
	inBlock := false
	for _, raw := range strings.Split(text, "\n") {
		var line string
		line, inBlock = stripLine(raw, inBlock)
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// stripLine удаляет комментарии из одной строки. Параметр inBlock
// показывает, начинается ли строка внутри блочного комментария /* */.
func stripLine(line string, inBlock bool) (string, bool) {
	var b strings.Builder
	for line != "" {
		if inBlock {
			end := strings.Index(line, "*/")
			if end < 0 {
				return b.String(), true
			}
			line, inBlock = line[end+markerLen:], false
			continue
		}
		start := commentStart(line)
		if start < 0 {
			b.WriteString(line)
			break
		}
		b.WriteString(line[:start])
		if strings.HasPrefix(line[start:], "//") {
			break
		}
		line, inBlock = line[start+markerLen:], true
	}
	return b.String(), inBlock
}

// commentStart возвращает позицию начала комментария // или /* в строке.
// Комментарий должен стоять в начале строки или после пробела, чтобы
// пути вида /a//b не считались комментариями. Возвращает -1, если его нет.
func commentStart(line string) int {
	for i := 0; i+1 < len(line); i++ {
		if line[i] != '/' || (line[i+1] != '/' && line[i+1] != '*') {
			continue
		}
		if i == 0 || line[i-1] == ' ' || line[i-1] == '\t' {
			return i
		}
	}
	return -1
}
