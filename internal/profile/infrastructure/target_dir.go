package infrastructure

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/united-software-platform/extruder/internal/profile/application"
)

// TargetDir реализует контракт целевого каталога слоя приложения.
var _ application.TargetDirectory = TargetDir{}

// probeName — имя файла-пробы, которым проверяется доступность каталога на запись. Имя скрытое:
// проба живёт в целевом каталоге доли секунды, но попасть на глаза пользователю не должна.
const probeName = ".extruder-write-probe"

// probeMode — права файла-пробы: читать и писать только владельцу.
const probeMode = 0o600

// TargetDir — работа с целевым каталогом через файловую систему. Состояния нет: путь приходит
// параметром каждой операции.
type TargetDir struct{}

// NewTargetDir возвращает реализацию контракта целевого каталога.
func NewTargetDir() application.TargetDirectory {
	return TargetDir{}
}

// EnsureWritable проверяет, что каталог существует, является каталогом и доступен на запись.
// Доступность проверяется пробной записью, а не битами прав: под числовым uid образа владелец
// смонтированного каталога почти наверняка чужой, и биты врут в обе стороны; заодно так
// обнаруживается монтирование только на чтение. Файл-проба удаляется немедленно, и неудача
// удаления — тоже отказ: иначе проба осталась бы в каталоге пользователя.
func (TargetDir) EnsureWritable(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("%w: каталог %q недоступен: %w", application.ErrTargetDirUnsuitable, dir, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("%w: %q — не каталог", application.ErrTargetDirUnsuitable, dir)
	}

	probe := filepath.Join(dir, probeName)

	file, err := os.OpenFile(probe, os.O_CREATE|os.O_EXCL|os.O_WRONLY, probeMode)
	if err != nil {
		return fmt.Errorf(
			"%w: в каталог %q нельзя писать: %w", application.ErrTargetDirUnsuitable, dir, err,
		)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf(
			"%w: файл-проба %q не закрыт: %w", application.ErrTargetDirUnsuitable, probe, err,
		)
	}

	if err := os.Remove(probe); err != nil {
		return fmt.Errorf(
			"%w: файл-проба %q не удалён: %w", application.ErrTargetDirUnsuitable, probe, err,
		)
	}

	return nil
}

// Entries возвращает имена элементов верхнего уровня каталога. Вложенные каталоги не обходятся:
// гейт предъявляет пользователю верхний уровень, а не всё дерево.
func (TargetDir) Entries(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("чтение каталога %q: %w", dir, err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	return names, nil
}
