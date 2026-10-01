//go:build acceptance

package acceptance

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/united-software-platform/extruder/internal/generation/application"
	profile "github.com/united-software-platform/extruder/internal/profile/domain"
)

// filePerm — права файла, записываемого фейковым генератором.
const filePerm = 0o644

// errFakeGenerator — отказ фейкового генератора.
var errFakeGenerator = errors.New("фейковый генератор: профиль не поддержан")

// fakeGenerator — генератор для проверки самого каркаса: пишет минимальный проект с оснасткой,
// объявляющей цели `build`, `test` и `lint`. Живёт в файле теста, в бинарник не попадает.
//
// Без него при пустом фильтре настоящий генератор не вызывается ни разу, и весь код каркаса ушёл
// бы в изменения Ядра непроверенным: первое падение было бы непонятно чьим.
type fakeGenerator struct {
	// targets — цели, объявляемые в оснастке; nil означает все три свойства.
	targets []string
	// failing — цели, завершающиеся ошибкой.
	failing []string
	// sink — файл, в который каждая вызванная цель дописывает отметку. Файл лежит вне каталога
	// профиля: каталог прошедшего профиля удаляется, а отметки нужны после прогона.
	sink string
	// rejecting — каноническое имя профиля, на котором генератор отказывает.
	rejecting string
}

// Execute пишет минимальный проект в переданный каталог либо отказывает на заданном профиле.
func (g fakeGenerator) Execute(_ context.Context, in application.GenerateStructureInput) error {
	if g.rejecting != "" && in.Profile.String() == g.rejecting {
		return fmt.Errorf("%w: %s", errFakeGenerator, in.Profile)
	}

	readme := fmt.Sprintf("# %s\n\nПрофиль: %s\n", in.ProjectIdentifier, in.Profile)
	if err := os.WriteFile(filepath.Join(in.TargetDir, "README.md"), []byte(readme), filePerm); err != nil {
		return fmt.Errorf("запись README.md: %w", err)
	}

	path := filepath.Join(in.TargetDir, makefileName)
	if err := os.WriteFile(path, []byte(g.makefile(in.Profile)), filePerm); err != nil {
		return fmt.Errorf("запись %s: %w", makefileName, err)
	}

	return nil
}

// makefile собирает оснастку минимального проекта.
func (g fakeGenerator) makefile(p profile.Profile) string {
	targets := g.targets
	if targets == nil {
		targets = properties()
	}

	var text strings.Builder

	fmt.Fprintf(&text, ".PHONY: %s\n\n", strings.Join(targets, " "))

	for _, target := range targets {
		fmt.Fprintf(&text, "%s:\n", target)

		if g.sink != "" {
			fmt.Fprintf(&text, "\t@echo '%s %s' >> '%s'\n", target, p, g.sink)
		}

		if slices.Contains(g.failing, target) {
			text.WriteString("\t@exit 1\n")
		} else {
			text.WriteString("\t@true\n")
		}

		text.WriteString("\n")
	}

	return text.String()
}

// recorder — перехватчик вердиктов: подставляется вместо `*testing.T`, чтобы тест каркаса видел
// и неуспешные исходы, не падая вместе с ними.
type recorder struct {
	logs     []string
	verdicts []string
	subtests []string
	failed   bool
}

// Run выполняет проверку профиля на вложенном перехватчике и возвращает её успешность.
func (r *recorder) Run(name string, body func(reporter)) bool {
	child := &recorder{}
	r.subtests = append(r.subtests, name)

	body(child)

	r.logs = append(r.logs, child.logs...)
	r.verdicts = append(r.verdicts, child.verdicts...)

	if child.failed {
		r.failed = true

		return false
	}

	return true
}

// Logf запоминает строку прогона.
func (r *recorder) Logf(format string, args ...any) {
	r.logs = append(r.logs, fmt.Sprintf(format, args...))
}

// Errorf запоминает вердикт «не прошёл».
func (r *recorder) Errorf(format string, args ...any) {
	r.verdicts = append(r.verdicts, fmt.Sprintf(format, args...))
	r.failed = true
}

// contains сообщает, встречается ли подстрока хотя бы в одной из строк.
func contains(lines []string, substring string) bool {
	for _, line := range lines {
		if strings.Contains(line, substring) {
			return true
		}
	}

	return false
}

// singleProfileFilter — фильтр, берущий ровно один профиль `go_monolith_clean-architecture`.
func singleProfileFilter() implementedValues {
	return implementedValues{
		languages:     []profile.Language{profile.LanguageGo},
		projectTypes:  []profile.ProjectType{profile.ProjectTypeMonolith},
		architectures: []profile.Architecture{profile.ArchitectureClean},
	}
}

// singleProfileName — каноническое имя профиля, который берёт singleProfileFilter.
const singleProfileName = "go_monolith_clean-architecture"

func TestFakeGeneratorWritesProject(t *testing.T) {
	p, err := profile.NewProfile(
		profile.LanguageGo,
		profile.ProjectTypeMonolith,
		profile.ArchitectureClean,
		nil,
	)
	if err != nil {
		t.Fatalf("создание профиля вернуло ошибку: %v", err)
	}

	dir := t.TempDir()

	err = fakeGenerator{}.Execute(context.Background(), application.GenerateStructureInput{
		Profile:           p,
		ProjectIdentifier: projectIdentifier(p),
		TargetDir:         dir,
	})
	if err != nil {
		t.Fatalf("фейковый генератор вернул ошибку: %v", err)
	}

	for _, name := range []string{makefileName, "README.md"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("в каталоге профиля нет файла %s: %v", name, err)
		}
	}

	for _, target := range properties() {
		if err := hasTarget(dir, target); err != nil {
			t.Errorf("оснастка фейка не объявляет цель %q: %v", target, err)
		}
	}
}

func TestRunChecksAllThreeProperties(t *testing.T) {
	root := t.TempDir()
	sink := filepath.Join(t.TempDir(), "calls.txt")

	result := runOverFake(t, root, fakeGenerator{sink: sink}, singleProfileFilter())

	if result != (runResult{taken: 1, passed: 1}) {
		t.Fatalf("итог прогона %+v, ожидался один взятый и прошедший профиль", result)
	}

	calls, err := os.ReadFile(sink)
	if err != nil {
		t.Fatalf("чтение отметок о вызовах целей: %v", err)
	}

	for _, target := range properties() {
		if !strings.Contains(string(calls), target+" "+singleProfileName) {
			t.Errorf("цель %q не вызвана; отметки:\n%s", target, calls)
		}
	}
}

func TestRunRemovesDirectoryOfPassedProfile(t *testing.T) {
	root := t.TempDir()

	runOverFake(t, root, fakeGenerator{}, singleProfileFilter())

	if _, err := os.Stat(filepath.Join(root, singleProfileName)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("каталог прошедшего профиля не удалён: %v", err)
	}
}

func TestRunKeepsDirectoryOfFailedProfile(t *testing.T) {
	root := t.TempDir()
	reported := &recorder{}

	result := run(reported, root, fakeGenerator{failing: []string{"test"}}, singleProfileFilter())

	if result != (runResult{taken: 1, failed: 1}) {
		t.Fatalf("итог прогона %+v, ожидался один взятый и упавший профиль", result)
	}

	dir := filepath.Join(root, singleProfileName)
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("каталог упавшего профиля не сохранён: %v", err)
	}

	if !contains(reported.verdicts, dir) {
		t.Errorf("вердикт не называет каталог профиля: %v", reported.verdicts)
	}

	if !contains(reported.verdicts, `цель "test" завершилась ошибкой`) {
		t.Errorf("вердикт не называет упавшую цель: %v", reported.verdicts)
	}
}

func TestRunReportsMissingTarget(t *testing.T) {
	reported := &recorder{}

	result := run(
		reported,
		t.TempDir(),
		fakeGenerator{targets: []string{"build", "test"}},
		singleProfileFilter(),
	)

	if result != (runResult{taken: 1, failed: 1}) {
		t.Fatalf("итог прогона %+v, ожидался один взятый и упавший профиль", result)
	}

	if !contains(reported.verdicts, errTargetMissing.Error()) {
		t.Errorf("вердикт не называет причину «цель отсутствует»: %v", reported.verdicts)
	}

	if contains(reported.verdicts, `цель "lint" завершилась ошибкой`) {
		t.Errorf("отсутствие цели предъявлено как её провал: %v", reported.verdicts)
	}
}

func TestRunReportsGeneratorRejection(t *testing.T) {
	root := t.TempDir()
	reported := &recorder{}

	result := run(reported, root, fakeGenerator{rejecting: singleProfileName}, singleProfileFilter())

	if result != (runResult{taken: 1, failed: 1}) {
		t.Fatalf("итог прогона %+v, ожидался один взятый и упавший профиль", result)
	}

	if !contains(reported.verdicts, "отказ генератора") {
		t.Errorf("вердикт не называет причину «отказ генератора»: %v", reported.verdicts)
	}

	if contains(reported.verdicts, "цель") {
		t.Errorf("после отказа генератора проверялись свойства: %v", reported.verdicts)
	}
}

func TestRunChecksEveryProfileAfterFailure(t *testing.T) {
	filter := singleProfileFilter()
	filter.patterns = []profile.Pattern{profile.PatternDDD}

	reported := &recorder{}
	result := run(reported, t.TempDir(), fakeGenerator{rejecting: singleProfileName}, filter)

	if result != (runResult{taken: 2, passed: 1, failed: 1}) {
		t.Fatalf("итог прогона %+v, ожидались два взятых профиля: один прошёл, один упал", result)
	}

	want := []string{singleProfileName, singleProfileName + "_ddd"}
	if !slices.Equal(reported.subtests, want) {
		t.Errorf("проверены профили %v, ожидались %v", reported.subtests, want)
	}

	if !reported.failed {
		t.Error("прогон завершился успехом, ожидался неуспех при упавшем профиле")
	}

	if !contains(reported.logs, summaryLine(result)) {
		t.Errorf("сводка %q не напечатана; строки прогона: %v", summaryLine(result), reported.logs)
	}
}

func TestRunClearsRootOnRepeat(t *testing.T) {
	root := t.TempDir()

	runOverFake(t, root, fakeGenerator{}, singleProfileFilter())

	// Каталог упавшего профиля предыдущего прогона: он остался бы на диске и смешался
	// с деревьями текущего, если бы служебный каталог не чистился в начале прогона.
	stale := filepath.Join(root, "python_service_layered_cqrs")
	if err := os.MkdirAll(stale, dirPerm); err != nil {
		t.Fatalf("подготовка каталога предыдущего прогона вернула ошибку: %v", err)
	}

	runOverFake(t, root, fakeGenerator{}, singleProfileFilter())

	if _, err := os.Stat(stale); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("каталог предыдущего прогона не удалён: %v", err)
	}
}

// TestRunReportsThroughTesting проверяет предъявление вердиктов средствами testing: профиль
// получает собственный подтест. Остальные тесты каркаса идут через перехватчик, поэтому без этого
// теста путь, которым прогон предъявляет вердикты в действительности, не исполнялся бы ни разу —
// при пустом фильтре у настоящего прогона подтестов нет.
func TestRunReportsThroughTesting(t *testing.T) {
	root := t.TempDir()
	sink := filepath.Join(t.TempDir(), "calls.txt")

	result := run(testingReporter{t: t}, root, fakeGenerator{sink: sink}, singleProfileFilter())

	if result != (runResult{taken: 1, passed: 1}) {
		t.Fatalf("итог прогона %+v, ожидался один взятый и прошедший профиль", result)
	}

	calls, err := os.ReadFile(sink)
	if err != nil {
		t.Fatalf("чтение отметок о вызовах целей: %v", err)
	}

	for _, target := range properties() {
		if !strings.Contains(string(calls), target+" "+singleProfileName) {
			t.Errorf("цель %q не вызвана; отметки:\n%s", target, calls)
		}
	}
}

// runOverFake выполняет прогон над фейком и требует, чтобы он прошёл без вердиктов «не прошёл».
func runOverFake(
	t *testing.T,
	root string,
	generator fakeGenerator,
	filter implementedValues,
) runResult {
	t.Helper()

	reported := &recorder{}
	result := run(reported, root, generator, filter)

	if reported.failed {
		t.Fatalf("прогон над фейком дал вердикты «не прошёл»: %v", reported.verdicts)
	}

	return result
}
