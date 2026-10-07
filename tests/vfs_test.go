package tests

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"shellemu/src/vfs"
)

// deepVFS — путь к примеру VFS с несколькими уровнями вложенности.
const deepVFS = "../examples/vfs/deep"

// TestResolve проверяет перевод относительных путей в абсолютные.
func TestResolve(t *testing.T) {
	cases := []struct{ cwd, p, want string }{
		{"/", "home", "/home"},
		{"/home/user", "..", "/home"},
		{"/home/user", "../../..", "/"},
		{"/home", "/etc/./config.cfg", "/etc/config.cfg"},
		{"/var", "~", "/"},
		{"/var", "~/home/user", "/home/user"},
		{"/var", "", "/"},
	}
	for _, c := range cases {
		if got := vfs.Resolve(c.cwd, c.p); got != c.want {
			t.Errorf("Resolve(%q, %q) = %q, want %q", c.cwd, c.p, got, c.want)
		}
	}
}

// TestLoadAndLookup проверяет загрузку директории и поиск узлов.
func TestLoadAndLookup(t *testing.T) {
	fsys, err := vfs.Load(deepVFS)
	if err != nil {
		t.Fatal(err)
	}
	node, err := fsys.Lookup("/home/user/docs/report.txt")
	if err != nil || node.IsDir || len(node.Data) == 0 {
		t.Fatalf("Lookup report.txt = %v, %v", node, err)
	}
	if _, err := fsys.Lookup("/home/missing"); !errors.Is(err, vfs.ErrNotExist) {
		t.Errorf("missing path error = %v", err)
	}
	if _, err := fsys.Lookup("/etc/config.cfg/x"); !errors.Is(err, vfs.ErrNotDir) {
		t.Errorf("file as dir error = %v", err)
	}
	if dirs, files := fsys.Stats(); dirs == 0 || files == 0 {
		t.Errorf("Stats = %d, %d", dirs, files)
	}
}

// TestLoadErrors проверяет ошибки при неверном источнике VFS.
func TestLoadErrors(t *testing.T) {
	if _, err := vfs.Load("../examples/vfs/no_such_dir"); err == nil {
		t.Error("expected error for missing directory")
	}
	_, err := vfs.Load("../examples/vfs/minimal/hello.txt")
	if !errors.Is(err, vfs.ErrSourceType) {
		t.Errorf("file source error = %v", err)
	}
}

// TestSave проверяет, что сохранённая VFS совпадает с исходной.
func TestSave(t *testing.T) {
	fsys, err := vfs.Load(deepVFS)
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(t.TempDir(), "out", "deep")
	if err := fsys.Save(target); err != nil {
		t.Fatal(err)
	}
	want, _ := os.ReadFile(filepath.Join(deepVFS, "etc", "config.cfg"))
	got, err := os.ReadFile(filepath.Join(target, "etc", "config.cfg"))
	if err != nil || string(got) != string(want) {
		t.Errorf("saved file = %q, %v", got, err)
	}
	if err := fsys.Save(target); !errors.Is(err, vfs.ErrExist) {
		t.Errorf("second Save error = %v", err)
	}
}
