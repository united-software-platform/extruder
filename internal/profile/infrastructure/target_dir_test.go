package infrastructure

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/united-software-platform/extruder/internal/profile/application"
)

func TestEnsureWritableAcceptsWritableDir(t *testing.T) {
	dir := t.TempDir()

	if err := NewTargetDir().EnsureWritable(dir); err != nil {
		t.Fatalf("пригодный каталог отвергнут: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("чтение каталога вернуло ошибку: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("проверка оставила в каталоге %d элементов, ожидался пустой каталог", len(entries))
	}
}

func TestEnsureWritableKeepsExistingContent(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "README.md")

	if err := os.WriteFile(existing, []byte("содержимое"), 0o600); err != nil {
		t.Fatalf("подготовка файла вернула ошибку: %v", err)
	}

	if err := NewTargetDir().EnsureWritable(dir); err != nil {
		t.Fatalf("пригодный каталог отвергнут: %v", err)
	}

	names, err := NewTargetDir().Entries(dir)
	if err != nil {
		t.Fatalf("перечисление каталога вернуло ошибку: %v", err)
	}

	if want := []string{"README.md"}; !slices.Equal(names, want) {
		t.Errorf("содержимое каталога %v, ожидалось %v: проверка оставила след", names, want)
	}
}

func TestEnsureWritableRejectsMissingDir(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "не-смонтирован")

	err := NewTargetDir().EnsureWritable(missing)
	if err == nil {
		t.Fatal("несуществующий каталог принят, ожидался отказ")
	}

	if !errors.Is(err, application.ErrTargetDirUnsuitable) {
		t.Errorf("ошибка %v не сравнима с ErrTargetDirUnsuitable", err)
	}
}

func TestEnsureWritableRejectsFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "файл")

	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatalf("подготовка файла вернула ошибку: %v", err)
	}

	err := NewTargetDir().EnsureWritable(path)
	if err == nil {
		t.Fatal("файл принят как каталог, ожидался отказ")
	}

	if !errors.Is(err, application.ErrTargetDirUnsuitable) {
		t.Errorf("ошибка %v не сравнима с ErrTargetDirUnsuitable", err)
	}
}

func TestEnsureWritableRejectsReadOnlyDir(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("прогон от root: права каталога на запись не влияют")
	}

	dir := filepath.Join(t.TempDir(), "только-чтение")
	if err := os.Mkdir(dir, 0o500); err != nil {
		t.Fatalf("подготовка каталога вернула ошибку: %v", err)
	}

	err := NewTargetDir().EnsureWritable(dir)
	if err == nil {
		t.Fatal("каталог без права записи принят, ожидался отказ")
	}

	if !errors.Is(err, application.ErrTargetDirUnsuitable) {
		t.Errorf("ошибка %v не сравнима с ErrTargetDirUnsuitable", err)
	}
}

func TestEntriesListsTopLevelOnly(t *testing.T) {
	dir := t.TempDir()

	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o700); err != nil {
		t.Fatalf("подготовка служебного каталога вернула ошибку: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "cmd", "extruder"), 0o700); err != nil {
		t.Fatalf("подготовка вложенного каталога вернула ошибку: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), nil, 0o600); err != nil {
		t.Fatalf("подготовка файла вернула ошибку: %v", err)
	}

	names, err := NewTargetDir().Entries(dir)
	if err != nil {
		t.Fatalf("перечисление каталога вернуло ошибку: %v", err)
	}

	want := []string{".git", ".gitignore", "cmd"}
	if !slices.Equal(names, want) {
		t.Errorf("содержимое каталога %v, ожидалось %v", names, want)
	}
}

func TestEntriesOnEmptyDir(t *testing.T) {
	names, err := NewTargetDir().Entries(t.TempDir())
	if err != nil {
		t.Fatalf("перечисление пустого каталога вернуло ошибку: %v", err)
	}

	if len(names) != 0 {
		t.Errorf("содержимое пустого каталога %v, ожидалось пустое", names)
	}
}

func TestEntriesRejectsMissingDir(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "не-смонтирован")

	if _, err := NewTargetDir().Entries(missing); err == nil {
		t.Fatal("перечисление несуществующего каталога вернуло успех, ожидался отказ")
	}
}
