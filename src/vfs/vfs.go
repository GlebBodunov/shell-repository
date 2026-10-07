// Package vfs реализует виртуальную файловую систему, которая целиком
// хранится в памяти. Источником VFS служит директория на диске.
package vfs

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// Права доступа по умолчанию для новых узлов VFS.
const (
	defaultDirMode  fs.FileMode = 0o755
	defaultFileMode fs.FileMode = 0o644
)

// Ошибки операций с VFS.
var (
	ErrNotExist   = errors.New("no such file or directory")
	ErrNotDir     = errors.New("not a directory")
	ErrIsDir      = errors.New("is a directory")
	ErrExist      = errors.New("file exists")
	ErrSourceType = errors.New("VFS source must be a directory")
	ErrRoot       = errors.New("cannot remove the root directory")
)

// Node — файл или директория VFS.
type Node struct {
	Name     string
	IsDir    bool
	Mode     fs.FileMode
	Data     []byte
	children map[string]*Node
}

// FS — виртуальная файловая система с корневой директорией "/".
type FS struct {
	root *Node
}

// New создаёт пустую VFS, содержащую только корневую директорию.
func New() *FS {
	return &FS{root: newDir("/", defaultDirMode)}
}

// newDir создаёт узел-директорию.
func newDir(name string, mode fs.FileMode) *Node {
	return &Node{Name: name, IsDir: true, Mode: mode, children: map[string]*Node{}}
}

// Children возвращает содержимое директории, отсортированное по имени.
func (n *Node) Children() []*Node {
	list := make([]*Node, 0, len(n.children))
	for _, child := range n.children {
		list = append(list, child)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	return list
}

// Resolve переводит путь p относительно текущей директории cwd в
// абсолютный путь VFS. Символ ~ обозначает корень VFS.
func Resolve(cwd, p string) string {
	switch {
	case p == "" || p == "~":
		return "/"
	case strings.HasPrefix(p, "~/"):
		return path.Clean("/" + p[1:])
	case strings.HasPrefix(p, "/"):
		return path.Clean(p)
	default:
		return path.Clean(path.Join(cwd, p))
	}
}

// Lookup находит узел по абсолютному пути VFS.
func (f *FS) Lookup(p string) (*Node, error) {
	node := f.root
	for _, part := range splitPath(p) {
		if !node.IsDir {
			return nil, ErrNotDir
		}
		child, ok := node.children[part]
		if !ok {
			return nil, ErrNotExist
		}
		node = child
	}
	return node, nil
}

// Stats возвращает количество директорий (без корня) и файлов в VFS.
func (f *FS) Stats() (dirs, files int) {
	var walk func(n *Node)
	walk = func(n *Node) {
		for _, child := range n.children {
			if child.IsDir {
				dirs++
				walk(child)
			} else {
				files++
			}
		}
	}
	walk(f.root)
	return dirs, files
}

// Remove удаляет файл или директорию по абсолютному пути VFS.
// Директория удаляется только при recursive, равном true.
// Изменения выполняются только в памяти.
func (f *FS) Remove(p string, recursive bool) error {
	p = path.Clean("/" + p)
	if p == "/" {
		return ErrRoot
	}
	parent, err := f.Lookup(path.Dir(p))
	if err != nil {
		return err
	}
	if !parent.IsDir {
		return ErrNotDir
	}
	node, ok := parent.children[path.Base(p)]
	if !ok {
		return ErrNotExist
	}
	if node.IsDir && !recursive {
		return ErrIsDir
	}
	delete(parent.children, node.Name)
	return nil
}

// splitPath разбивает абсолютный путь на имена компонентов.
func splitPath(p string) []string {
	clean := strings.Trim(path.Clean("/"+p), "/")
	if clean == "" {
		return nil
	}
	return strings.Split(clean, "/")
}

// Load загружает в память содержимое директории dir. Исходные данные
// на диске не изменяются. Символьные ссылки и специальные файлы пропускаются.
func Load(dir string) (*FS, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, ErrSourceType
	}
	root := newDir("/", info.Mode().Perm())
	if err := loadDir(root, dir); err != nil {
		return nil, err
	}
	return &FS{root: root}, nil
}

// loadDir рекурсивно читает директорию диска в узел VFS.
func loadDir(node *Node, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		child, err := loadEntry(entry, filepath.Join(dir, entry.Name()))
		if err != nil {
			return err
		}
		if child != nil {
			node.children[child.Name] = child
		}
	}
	return nil
}

// loadEntry читает один элемент директории. Возвращает nil для
// символьных ссылок и специальных файлов.
func loadEntry(entry fs.DirEntry, full string) (*Node, error) {
	info, err := entry.Info()
	if err != nil {
		return nil, err
	}
	if entry.IsDir() {
		child := newDir(entry.Name(), info.Mode().Perm())
		return child, loadDir(child, full)
	}
	if !info.Mode().IsRegular() {
		return nil, nil
	}
	data, err := os.ReadFile(full)
	if err != nil {
		return nil, err
	}
	return &Node{Name: entry.Name(), Mode: info.Mode().Perm(), Data: data}, nil
}

// Save сохраняет состояние VFS на диск в исходном формате — в виде
// директории dir. Директория dir не должна существовать, недостающие
// родительские директории создаются автоматически.
func (f *FS) Save(dir string) error {
	if _, err := os.Lstat(dir); err == nil {
		return ErrExist
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dir), defaultDirMode); err != nil {
		return err
	}
	return saveNode(f.root, dir)
}

// saveNode рекурсивно записывает узел VFS на диск.
func saveNode(node *Node, target string) error {
	if !node.IsDir {
		return os.WriteFile(target, node.Data, modeOr(node.Mode, defaultFileMode))
	}
	if err := os.Mkdir(target, modeOr(node.Mode, defaultDirMode)); err != nil {
		return err
	}
	for _, child := range node.Children() {
		if err := saveNode(child, filepath.Join(target, child.Name)); err != nil {
			return err
		}
	}
	return nil
}

// modeOr возвращает mode или значение по умолчанию, если права не заданы.
func modeOr(mode, fallback fs.FileMode) fs.FileMode {
	if mode == 0 {
		return fallback
	}
	return mode
}
