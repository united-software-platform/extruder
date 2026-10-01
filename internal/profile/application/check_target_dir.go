package application

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/united-software-platform/extruder/internal/profile/domain"
)

// Текст вопроса о продолжении на непустом целевом каталоге. Вопрос один на всё содержимое:
// поштучный опрос на десятках файлов приводит к бездумному согласию.
const questionContinueOnNonEmpty = "Продолжить? (да/нет)"

// CheckTargetDirInput — входной контракт сценария проверки целевого каталога. Путь приходит
// параметром: точку монтирования подставляет точка входа, а тест — свой временный каталог.
type CheckTargetDirInput struct {
	Dir string
}

// CheckTargetDirOutput — выходной контракт сценария: продолжать ли запуск. Отказ продолжать —
// решение человека, а не ошибка, поэтому он приходит значением, а не ошибкой.
type CheckTargetDirOutput struct {
	Proceed bool
}

// CheckTargetDirUseCase — сценарий проверки целевого каталога: предусловие опроса, отвечающее
// на вопрос «можно ли начинать спрашивать». В целевой каталог сценарий ничего не пишет и о плане
// генерации не знает.
type CheckTargetDirUseCase interface {
	Execute(in CheckTargetDirInput) (CheckTargetDirOutput, error)
}

// checkTargetDir — реализация сценария проверки целевого каталога.
type checkTargetDir struct {
	target   TargetDirectory
	prompter Prompter
}

// NewCheckTargetDir возвращает сценарий проверки целевого каталога. Сценарий зависит только
// от контрактов, объявленных этим слоем, а не от их реализаций.
func NewCheckTargetDir(target TargetDirectory, prompter Prompter) CheckTargetDirUseCase {
	return checkTargetDir{target: target, prompter: prompter}
}

// Execute проверяет целевой каталог до первого вопроса опроса: непригодный каталог отвергается
// ошибкой с причиной, пустой с точностью до служебных каталогов проходит молча, непустой
// предъявляется перечнем путей и одним вопросом с умолчанием «нет».
func (c checkTargetDir) Execute(in CheckTargetDirInput) (CheckTargetDirOutput, error) {
	if err := c.target.EnsureWritable(in.Dir); err != nil {
		return CheckTargetDirOutput{}, err
	}

	entries, err := c.target.Entries(in.Dir)
	if err != nil {
		return CheckTargetDirOutput{}, fmt.Errorf(
			"%w: содержимое каталога %q не прочитано: %w", ErrTargetDirUnsuitable, in.Dir, err,
		)
	}

	blocking := blockingPaths(in.Dir, entries)
	if len(blocking) == 0 {
		return CheckTargetDirOutput{Proceed: true}, nil
	}

	answer, err := c.prompter.ReadLine(nonEmptyPrompt(in.Dir, blocking))
	if err != nil {
		return CheckTargetDirOutput{}, err
	}

	return CheckTargetDirOutput{Proceed: confirms(answer)}, nil
}

// blockingPaths оставляет от содержимого каталога только то, что мешает считать его пустым,
// и приводит имена к путям верхнего уровня. Решение «мешает ли элемент» принимает домен.
func blockingPaths(dir string, entries []string) []string {
	blocking := make([]string, 0, len(entries))

	for _, entry := range entries {
		if domain.BlocksEmptiness(entry) {
			blocking = append(blocking, filepath.Join(dir, entry))
		}
	}

	return blocking
}

// nonEmptyPrompt собирает подсказку вопроса о продолжении: перечень путей верхнего уровня
// и один общий вопрос.
func nonEmptyPrompt(dir string, blocking []string) string {
	lines := make([]string, 0, len(blocking)+4)
	lines = append(lines, fmt.Sprintf("Целевой каталог %s не пуст:", dir))

	for _, path := range blocking {
		lines = append(lines, "  "+path)
	}

	lines = append(lines, "", "Перечисленное может быть перезаписано.", questionContinueOnNonEmpty)

	return strings.Join(lines, "\n")
}
