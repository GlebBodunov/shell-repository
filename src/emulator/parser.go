// Package emulator реализует эмулятор командной оболочки UNIX-подобной ОС.
package emulator

import (
	"errors"
	"strings"
	"unicode"
)

// ErrUnterminatedQuote возвращается, если в строке не закрыта кавычка.
var ErrUnterminatedQuote = errors.New("unterminated quote")

// ErrBadSubstitution возвращается при некорректной конструкции ${...}.
var ErrBadSubstitution = errors.New("bad substitution")

// lexer хранит состояние разбора одной строки команды.
type lexer struct {
	src    []rune
	pos    int
	lookup func(string) string
	words  []string
	cur    strings.Builder
	inWord bool
}

// Parse разбивает строку на слова с учётом одинарных и двойных кавычек,
// экранирования обратной косой чертой и раскрытия переменных окружения
// вида $NAME и ${NAME}. Значения переменных берутся из функции lookup.
func Parse(line string, lookup func(string) string) ([]string, error) {
	l := &lexer{src: []rune(line), lookup: lookup}
	for l.pos < len(l.src) {
		if err := l.step(); err != nil {
			return nil, err
		}
	}
	l.flush()
	return l.words, nil
}

// step обрабатывает очередной символ строки вне кавычек.
func (l *lexer) step() error {
	r := l.src[l.pos]
	switch {
	case unicode.IsSpace(r):
		l.flush()
		l.pos++
	case r == '\'':
		return l.singleQuoted()
	case r == '"':
		return l.doubleQuoted()
	case r == '$':
		return l.variable()
	case r == '\\':
		l.escaped()
	default:
		l.add(string(r))
		l.pos++
	}
	return nil
}

// add дописывает текст к текущему слову.
func (l *lexer) add(s string) {
	if s != "" {
		l.inWord = true
	}
	l.cur.WriteString(s)
}

// flush завершает текущее слово и добавляет его в результат.
func (l *lexer) flush() {
	if l.inWord {
		l.words = append(l.words, l.cur.String())
	}
	l.cur.Reset()
	l.inWord = false
}

// escaped добавляет символ, следующий за обратной косой чертой, без изменений.
func (l *lexer) escaped() {
	l.pos++
	if l.pos < len(l.src) {
		l.add(string(l.src[l.pos]))
		l.pos++
	}
}

// singleQuoted читает текст в одинарных кавычках без раскрытия переменных.
func (l *lexer) singleQuoted() error {
	end := indexRune(l.src, l.pos+1, '\'')
	if end < 0 {
		return ErrUnterminatedQuote
	}
	l.inWord = true
	l.add(string(l.src[l.pos+1 : end]))
	l.pos = end + 1
	return nil
}

// doubleQuoted читает текст в двойных кавычках с раскрытием переменных.
func (l *lexer) doubleQuoted() error {
	l.inWord = true
	l.pos++
	for l.pos < len(l.src) {
		r := l.src[l.pos]
		switch r {
		case '"':
			l.pos++
			return nil
		case '$':
			if err := l.variable(); err != nil {
				return err
			}
		case '\\':
			l.escaped()
		default:
			l.add(string(r))
			l.pos++
		}
	}
	return ErrUnterminatedQuote
}

// variable раскрывает переменную окружения, начинающуюся с символа $.
// Одиночный символ $ без имени переменной остаётся в тексте как есть.
func (l *lexer) variable() error {
	l.pos++
	if l.pos < len(l.src) && l.src[l.pos] == '{' {
		return l.bracedVariable()
	}
	start := l.pos
	for l.pos < len(l.src) && isNameRune(l.src[l.pos]) {
		l.pos++
	}
	if start == l.pos {
		l.add("$")
		return nil
	}
	l.add(l.lookup(string(l.src[start:l.pos])))
	return nil
}

// bracedVariable раскрывает переменную в форме ${NAME}.
func (l *lexer) bracedVariable() error {
	end := indexRune(l.src, l.pos+1, '}')
	if end < 0 {
		return ErrBadSubstitution
	}
	name := string(l.src[l.pos+1 : end])
	if name == "" {
		return ErrBadSubstitution
	}
	l.add(l.lookup(name))
	l.pos = end + 1
	return nil
}

// indexRune ищет символ target в src, начиная с позиции from.
// Возвращает -1, если символ не найден.
func indexRune(src []rune, from int, target rune) int {
	for i := from; i < len(src); i++ {
		if src[i] == target {
			return i
		}
	}
	return -1
}

// isNameRune сообщает, может ли символ входить в имя переменной окружения.
func isNameRune(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
