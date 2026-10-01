//go:build acceptance

package acceptance

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/united-software-platform/extruder/internal/generation/application"
	profile "github.com/united-software-platform/extruder/internal/profile/domain"
)

// runRootPath — служебный каталог прогона относительно корня репозитория. `.gocache` уже
// игнорируется системой контроля версий и уже служит каталогом кэшей целей Go.
const runRootPath = ".gocache/acceptance"

// dirPerm — права каталога профиля.
const dirPerm = 0o755

// makefileName — файл оснастки сгенерированного проекта, через цели которого проверяются свойства.
const makefileName = "Makefile"

// errTargetMissing — в оснастке сгенерированного проекта нет нужной цели. Отдельная причина
// вердикта: отсутствие цели и её провал — разные дефекты, и код возврата make их не различает.
var errTargetMissing = errors.New("цель отсутствует в оснастке")

// errRepositoryRootNotFound — корень репозитория не найден: над рабочим каталогом нет `go.mod`.
var errRepositoryRootNotFound = errors.New("корень репозитория не найден")

// properties возвращает цели оснастки, которыми проверяются три свойства сгенерированного
// проекта: проект собирается, его тест проходит, линтер чист. Что именно делает цель — дело
// сгенерированного проекта: прогон не знает ни языка профиля, ни инструментов этого языка.
func properties() []string {
	return []string{"build", "test", "lint"}
}

// runResult — итог прогона: сколько профилей взято, сколько прошло и сколько упало.
type runResult struct {
	taken  int
	passed int
	failed int
}

// reporter — то, через что прогон предъявляет охват и вердикты. В прогоне это `*testing.T`;
// тест каркаса подставляет перехватчик и потому проверяет сами вердикты, а не только их число.
type reporter interface {
	Run(name string, body func(reporter)) bool
	Logf(format string, args ...any)
	Errorf(format string, args ...any)
}

// testingReporter — предъявление вердиктов средствами testing: профиль получает собственный
// подтест, и его вердикт виден в выводе прогона отдельной строкой.
type testingReporter struct {
	t *testing.T
}

// Run выполняет проверку профиля отдельным подтестом и возвращает его успешность.
func (r testingReporter) Run(name string, body func(reporter)) bool {
	return r.t.Run(name, func(t *testing.T) {
		body(testingReporter{t: t})
	})
}

// Logf печатает строку прогона.
func (r testingReporter) Logf(format string, args ...any) {
	r.t.Helper()
	r.t.Logf(format, args...)
}

// Errorf печатает вердикт «не прошёл» и помечает прогон неуспешным.
func (r testingReporter) Errorf(format string, args ...any) {
	r.t.Helper()
	r.t.Errorf(format, args...)
}

// run выполняет прогон целиком: очищает служебный каталог, отбирает профили фильтром, печатает
// охват, проверяет каждый взятый профиль и печатает итоговую сводку прогона. Падение профиля
// прогон не прерывает:
// на этапах раскладок падают семейства профилей, и семейство видно только на полном прогоне.
func run(
	r reporter,
	root string,
	generator application.GenerateStructureUseCase,
	filter implementedValues,
) runResult {
	var result runResult

	if err := prepareRunRoot(root); err != nil {
		r.Errorf("подготовка служебного каталога прогона: %v", err)

		return result
	}

	all := profile.All()
	taken := selectProfiles(all, filter)
	result.taken = len(taken)
	r.Logf("%s", coverageLine(len(taken), len(all)))

	for _, p := range taken {
		if checkProfile(r, root, generator, p) {
			result.passed++
		} else {
			result.failed++
		}
	}

	r.Logf("%s", summaryLine(result))

	return result
}

// checkProfile проверяет один профиль: генерирует проект в его каталог и проверяет три свойства.
// Каталог прошедшего профиля удаляется, каталог упавшего остаётся для разбора и назван
// в вердикте — диагностика идёт по дереву, а не по тексту сообщения.
func checkProfile(
	r reporter,
	root string,
	generator application.GenerateStructureUseCase,
	p profile.Profile,
) bool {
	return r.Run(p.String(), func(r reporter) {
		dir := filepath.Join(root, p.String())
		if err := os.MkdirAll(dir, dirPerm); err != nil {
			r.Errorf("профиль %s: не удалось создать каталог %s: %v", p, dir, err)

			return
		}

		in := application.GenerateStructureInput{
			Profile:           p,
			ProjectIdentifier: projectIdentifier(p),
			TargetDir:         dir,
		}

		if err := generator.Execute(context.Background(), in); err != nil {
			r.Errorf("профиль %s: отказ генератора: %v; каталог профиля — %s", p, err, dir)

			return
		}

		passed := true

		for _, target := range properties() {
			if !checkProperty(r, p, dir, target) {
				passed = false
			}
		}

		if !passed {
			return
		}

		if err := os.RemoveAll(dir); err != nil {
			r.Errorf("профиль %s: не удалось удалить каталог %s: %v", p, dir, err)
		}
	})
}

// checkProperty проверяет одно свойство сгенерированного проекта вызовом цели его оснастки.
// Свойство считается выполненным, если цель завершилась успешно; вывод упавшей цели предъявлен
// целиком — своего сообщения о причине у прогона нет.
func checkProperty(r reporter, p profile.Profile, dir, target string) bool {
	if err := hasTarget(dir, target); err != nil {
		r.Errorf("профиль %s: цель %q не вызвана: %v; каталог профиля — %s", p, target, err, dir)

		return false
	}

	command := exec.Command("make", target)
	command.Dir = dir

	output, err := command.CombinedOutput()
	if err != nil {
		r.Errorf(
			"профиль %s: цель %q завершилась ошибкой: %v; каталог профиля — %s\n%s",
			p, target, err, dir, output,
		)

		return false
	}

	return true
}

// hasTarget проверяет, объявлена ли цель в оснастке сгенерированного проекта. Признак —
// объявление цели в самом файле, а не код возврата make: у отсутствующей и у упавшей цели он
// одинаков, а вердикты у них разные.
func hasTarget(dir, target string) error {
	path := filepath.Join(dir, makefileName)

	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%w: в сгенерированном проекте нет файла %s", errTargetMissing, makefileName)
	}

	if err != nil {
		return fmt.Errorf("чтение %s: %w", path, err)
	}

	declaration := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(target) + `[ \t]*:`)
	if !declaration.Match(content) {
		return fmt.Errorf("%w: цель %q не объявлена в %s", errTargetMissing, target, makefileName)
	}

	return nil
}

// defaultRunRoot возвращает штатный служебный каталог прогона. Корень репозитория ищется
// по `go.mod`: `go test` выполняет пакет из его собственного каталога, а служебный каталог живёт
// в корне рядом с кэшами целей Go.
func defaultRunRoot() (string, error) {
	root, err := repositoryRoot()
	if err != nil {
		return "", err
	}

	return filepath.Join(root, runRootPath), nil
}

// repositoryRoot ищет корень репозитория подъёмом от рабочего каталога до каталога с `go.mod`.
func repositoryRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("рабочий каталог: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("%w: над %s нет go.mod", errRepositoryRootNotFound, dir)
		}

		dir = parent
	}
}

// prepareRunRoot очищает служебный каталог прогона и создаёт его заново: деревья предыдущего
// прогона не должны смешиваться с деревьями текущего.
func prepareRunRoot(root string) error {
	if err := os.RemoveAll(root); err != nil {
		return fmt.Errorf("очистка %s: %w", root, err)
	}

	if err := os.MkdirAll(root, dirPerm); err != nil {
		return fmt.Errorf("создание %s: %w", root, err)
	}

	return nil
}
