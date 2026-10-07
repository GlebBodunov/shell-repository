package emulator

import (
	"bytes"
	"fmt"
	"strings"

	"shellemu/src/vfs"
)

// wcCounts — результат подсчёта строк, слов и байтов.
type wcCounts struct {
	lines, words, bytes int
}

// cmdWhoami выводит имя текущего пользователя реальной ОС.
func cmdWhoami(s *Shell, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("extra operand '%s'", args[0])
	}
	fmt.Fprintln(s.out, s.user)
	return nil
}

// cmdRev выводит строки файлов VFS, переворачивая каждую строку посимвольно.
func cmdRev(s *Shell, args []string) error {
	if len(args) == 0 {
		return errMissingOperand
	}
	for _, p := range args {
		data, err := s.readFile(p)
		if err != nil {
			s.printErr("rev", p, err)
			continue
		}
		for _, line := range splitLines(string(data)) {
			fmt.Fprintln(s.out, reverse(line))
		}
	}
	return nil
}

// cmdWc выводит число строк, слов и байтов в файлах VFS.
// Опции -l, -w и -c ограничивают вывод отдельными счётчиками.
func cmdWc(s *Shell, args []string) error {
	opts, paths, err := parseFlags(args, "lwc")
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		return errMissingOperand
	}
	if len(opts) == 0 {
		opts = map[rune]bool{'l': true, 'w': true, 'c': true}
	}
	var total wcCounts
	for _, p := range paths {
		data, err := s.readFile(p)
		if err != nil {
			s.printErr("wc", p, err)
			continue
		}
		c := countText(data)
		total = wcCounts{total.lines + c.lines, total.words + c.words, total.bytes + c.bytes}
		fmt.Fprintln(s.out, formatCounts(c, opts, p))
	}
	if len(paths) > singleOperand {
		fmt.Fprintln(s.out, formatCounts(total, opts, "total"))
	}
	return nil
}

// readFile возвращает содержимое файла VFS по пути относительно cwd.
func (s *Shell) readFile(p string) ([]byte, error) {
	node, err := s.fs.Lookup(vfs.Resolve(s.cwd, p))
	if err != nil {
		return nil, err
	}
	if node.IsDir {
		return nil, vfs.ErrIsDir
	}
	return node.Data, nil
}

// countText считает строки (символы перевода строки), слова и байты.
func countText(data []byte) wcCounts {
	return wcCounts{
		lines: bytes.Count(data, []byte("\n")),
		words: len(strings.Fields(string(data))),
		bytes: len(data),
	}
}

// formatCounts форматирует строку вывода wc с выбранными счётчиками.
func formatCounts(c wcCounts, opts map[rune]bool, name string) string {
	var b strings.Builder
	if opts['l'] {
		fmt.Fprintf(&b, "%7d ", c.lines)
	}
	if opts['w'] {
		fmt.Fprintf(&b, "%7d ", c.words)
	}
	if opts['c'] {
		fmt.Fprintf(&b, "%7d ", c.bytes)
	}
	b.WriteString(name)
	return b.String()
}

// splitLines разбивает текст на строки без завершающих символов перевода.
func splitLines(text string) []string {
	if text == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(text, "\n"), "\n")
}

// reverse переворачивает строку посимвольно с учётом UTF-8.
func reverse(line string) string {
	runes := []rune(line)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
