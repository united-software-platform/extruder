package infra_test

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/united-software-platform/extruder/internal/profile/domain"
	"github.com/united-software-platform/extruder/internal/profile/infra"
)

// run выполняет команду и возвращает код возврата вместе с обоими потоками.
func run(args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := infra.Run(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func fullArgs() []string {
	return []string{
		"--language", "go",
		"--project-type", "service",
		"--architecture", "layered",
		"--patterns", "ddd,cqrs",
	}
}

func TestRunAcceptsFullProfile(t *testing.T) {
	code, stdout, stderr := run(fullArgs()...)
	if code != 0 {
		t.Fatalf("код возврата %d, ожидался 0; stderr: %s", code, stderr)
	}
	for _, dimension := range []domain.Dimension{
		domain.DimensionLanguage,
		domain.DimensionProjectType,
		domain.DimensionArchitecture,
		domain.DimensionPatterns,
	} {
		if !strings.Contains(stdout, string(dimension)) {
			t.Errorf("вывод не содержит измерение %q: %s", string(dimension), stdout)
		}
	}
	if stderr != "" {
		t.Errorf("поток ошибок не пуст при успехе: %s", stderr)
	}
}

func TestRunIgnoresFlagOrder(t *testing.T) {
	_, direct, _ := run(fullArgs()...)
	_, reversed, _ := run(
		"--patterns", "ddd,cqrs",
		"--architecture", "layered",
		"--project-type", "service",
		"--language", "go",
	)
	if direct != reversed {
		t.Errorf("порядок флагов изменил результат:\n%s\nпротив\n%s", direct, reversed)
	}
}

func TestRunAcceptsProfileWithoutPatterns(t *testing.T) {
	code, stdout, stderr := run(
		"--language", "go", "--project-type", "service", "--architecture", "layered",
	)
	if code != 0 {
		t.Fatalf("код возврата %d, ожидался 0; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, string(domain.DimensionPatterns)) {
		t.Errorf("пустой набор паттернов не предъявлен: %s", stdout)
	}
}

func TestRunReportsMissingDimensions(t *testing.T) {
	code, stdout, stderr := run("--language", "go")
	if code == 0 {
		t.Fatal("неполный профиль принят")
	}
	if stdout != "" {
		t.Errorf("поток вывода не пуст при отказе: %s", stdout)
	}
	for _, dimension := range []domain.Dimension{domain.DimensionProjectType, domain.DimensionArchitecture} {
		if !strings.Contains(stderr, string(dimension)) {
			t.Errorf("диагностика не называет измерение %q: %s", string(dimension), stderr)
		}
	}
}

func TestRunReportsInvalidValueToStderrOnly(t *testing.T) {
	code, stdout, stderr := run(
		"--language", "rust", "--project-type", "service", "--architecture", "layered",
	)
	if code == 0 {
		t.Fatal("неизвестный язык принят")
	}
	if stdout != "" {
		t.Errorf("поток вывода не пуст при отказе: %s", stdout)
	}
	if !strings.Contains(stderr, "rust") || !strings.Contains(stderr, "go") {
		t.Errorf("диагностика не содержит значение и перечень допустимых: %s", stderr)
	}
}

func TestRunRejectsDuplicatePattern(t *testing.T) {
	code, _, stderr := run(append(fullArgs()[:6], "--patterns", "ddd,ddd")...)
	if code == 0 {
		t.Fatalf("повторённый паттерн принят; stderr: %s", stderr)
	}
}

func TestRunPrintsHelpOnRequest(t *testing.T) {
	code, stdout, _ := run("--help")
	if code != 0 {
		t.Fatalf("код возврата %d, ожидался 0", code)
	}
	for _, dimension := range []domain.Dimension{
		domain.DimensionLanguage,
		domain.DimensionProjectType,
		domain.DimensionArchitecture,
		domain.DimensionPatterns,
	} {
		if !strings.Contains(stdout, string(dimension)) {
			t.Errorf("справка не называет измерение %q", string(dimension))
		}
	}
	for _, obligation := range []string{"обязательно", "необязательно"} {
		if !strings.Contains(stdout, obligation) {
			t.Errorf("справка не сообщает обязательность (%q)", obligation)
		}
	}
	for _, value := range append(domain.AllowedValues(domain.DimensionLanguage),
		domain.AllowedValues(domain.DimensionPatterns)...) {
		if !strings.Contains(stdout, value) {
			t.Errorf("справка не называет допустимое значение %q", value)
		}
	}
}

func TestRunWithoutArgumentsFailsAndShowsHelp(t *testing.T) {
	code, stdout, stderr := run()
	if code == 0 {
		t.Fatal("вызов без аргументов завершился успехом")
	}
	if stdout != "" {
		t.Errorf("поток вывода не пуст при отказе: %s", stdout)
	}
	if !strings.Contains(stderr, string(domain.DimensionLanguage)) {
		t.Errorf("справка не выведена: %s", stderr)
	}
}

// listTree возвращает отсортированный перечень путей внутри каталога.
func listTree(t *testing.T, root string) []string {
	t.Helper()
	var paths []string
	err := filepath.Walk(root, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		t.Fatalf("обход каталога: %v", err)
	}
	sort.Strings(paths)
	return paths
}

func TestRunDoesNotTouchFilesystem(t *testing.T) {
	root := t.TempDir()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("текущий каталог: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("переход в каталог: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })

	before := listTree(t, root)
	run(fullArgs()...)
	run("--language", "rust", "--project-type", "service", "--architecture", "layered")
	run()
	after := listTree(t, root)

	if strings.Join(before, "\n") != strings.Join(after, "\n") {
		t.Errorf("состав каталога изменился:\nдо:  %v\nпосле: %v", before, after)
	}
}
