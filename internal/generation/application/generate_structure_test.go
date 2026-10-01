package application

import (
	"context"
	"errors"
	"os"
	"testing"

	profile "github.com/united-software-platform/extruder/internal/profile/domain"
)

func TestGenerateStructureStubRejects(t *testing.T) {
	p, err := profile.NewProfile(
		profile.LanguageGo,
		profile.ProjectTypeMonolith,
		profile.ArchitectureClean,
		nil,
	)
	if err != nil {
		t.Fatalf("создание профиля вернуло ошибку: %v", err)
	}

	targetDir := t.TempDir()

	err = NewGenerateStructure().Execute(context.Background(), GenerateStructureInput{
		Profile:           p,
		ProjectIdentifier: "acceptance-" + p.String(),
		TargetDir:         targetDir,
	})
	if err == nil {
		t.Fatal("заглушка вернула успех, ожидался отказ")
	}

	if !errors.Is(err, ErrNotImplemented) {
		t.Errorf("ошибка %v не сравнима с ErrNotImplemented", err)
	}

	entries, readErr := os.ReadDir(targetDir)
	if readErr != nil {
		t.Fatalf("чтение каталога %s вернуло ошибку: %v", targetDir, readErr)
	}

	if len(entries) != 0 {
		t.Errorf("заглушка записала в каталог %d элементов, ожидался пустой каталог", len(entries))
	}
}
