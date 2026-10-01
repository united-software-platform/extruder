package application

import (
	"context"
	"errors"
	"fmt"

	profile "github.com/united-software-platform/extruder/internal/profile/domain"
)

// ErrNotImplemented — генератор начальной структуры проекта ещё не реализован. Точку входа
// объявляет приёмочный прогон, а заполняют её изменения Ядра; до тех пор реализация-заглушка
// отказывает этой ошибкой. Отказ, а не успех: иначе первый же взятый в прогон профиль прошёл бы
// проверку свойств над пустым каталогом.
var ErrNotImplemented = errors.New("генерация начальной структуры проекта не реализована")

// GenerateStructureInput — входной контракт точки входа генератора: архитектурный профиль,
// идентификатор проекта и каталог, в который записывается начальная структура проекта.
// Контракт неизменяем и логики не содержит.
//
// Каталог приходит параметром, а не берётся из фиксированной точки монтирования: прогон пишет
// отдельный каталог на каждый профиль, а точку монтирования подставляет `cmd/extruder`.
type GenerateStructureInput struct {
	Profile           profile.Profile
	ProjectIdentifier string
	TargetDir         string
}

// GenerateStructureUseCase — точка входа генератора: единственная операция, записывающая
// начальную структуру проекта по заданному профилю. Выходного контракта нет — приёмочный прогон
// проверяет результат свойствами сгенерированного проекта и выход не читает.
type GenerateStructureUseCase interface {
	Execute(ctx context.Context, in GenerateStructureInput) error
}

// generateStructure — реализация-заглушка точки входа генератора.
type generateStructure struct{}

// NewGenerateStructure возвращает точку входа генератора. Вызывающая сторона получает реализацию
// конструктором, поэтому контракт компилируется вместе с ней, а не только со своими тестами.
func NewGenerateStructure() GenerateStructureUseCase {
	return generateStructure{}
}

// Execute отказывает ошибкой ErrNotImplemented и в переданный каталог ничего не пишет.
func (generateStructure) Execute(_ context.Context, in GenerateStructureInput) error {
	return fmt.Errorf("%w: профиль %s, каталог %q", ErrNotImplemented, in.Profile, in.TargetDir)
}
