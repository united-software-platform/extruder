//go:build acceptance

package acceptance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/united-software-platform/extruder/internal/generation/application"
	profile "github.com/united-software-platform/extruder/internal/profile/domain"
)

// TestAcceptance — приёмочный прогон: профили берутся из словаря, отбираются действующим фильтром
// реализованных значений и проверяются тремя свойствами. Пока фильтр пуст, прогон берёт ноль
// профилей и завершается успешно — и всё равно печатает охват.
func TestAcceptance(t *testing.T) {
	root, err := defaultRunRoot()
	if err != nil {
		t.Fatalf("служебный каталог прогона: %v", err)
	}

	run(testingReporter{t: t}, root, application.NewGenerateStructure(), implemented())
}

func TestImplementedFilterIsEmpty(t *testing.T) {
	taken := selectProfiles(profile.All(), implemented())
	if len(taken) != 0 {
		t.Errorf("действующий фильтр взял %d профилей, ожидался пустой отбор", len(taken))
	}
}

func TestSelectProfilesTakesProfileWithImplementedValues(t *testing.T) {
	filter := implementedValues{
		languages:     []profile.Language{profile.LanguageGo},
		projectTypes:  []profile.ProjectType{profile.ProjectTypeMonolith},
		architectures: []profile.Architecture{profile.ArchitectureClean},
	}

	taken := selectProfiles(profile.All(), filter)
	if len(taken) != 1 {
		t.Fatalf("взято %d профилей, ожидался ровно один", len(taken))
	}

	const want = "go_monolith_clean-architecture"
	if got := taken[0].String(); got != want {
		t.Errorf("взят профиль %s, ожидался %s", got, want)
	}
}

func TestSelectProfilesRejectsProfileWithUnimplementedPattern(t *testing.T) {
	filter := implementedValues{
		languages:     []profile.Language{profile.LanguageGo},
		projectTypes:  []profile.ProjectType{profile.ProjectTypeMonolith},
		architectures: []profile.Architecture{profile.ArchitectureClean},
		patterns:      []profile.Pattern{profile.PatternDDD},
	}

	taken := selectProfiles(profile.All(), filter)

	want := map[string]bool{
		"go_monolith_clean-architecture":     true,
		"go_monolith_clean-architecture_ddd": true,
	}

	if len(taken) != len(want) {
		t.Fatalf("взято %d профилей, ожидалось %d", len(taken), len(want))
	}

	for _, p := range taken {
		if !want[p.String()] {
			t.Errorf("взят профиль %s, значения его паттернов не все реализованы", p)
		}
	}
}

func TestSelectProfilesTakesNothingWithEmptyFilter(t *testing.T) {
	if taken := selectProfiles(profile.All(), implementedValues{}); len(taken) != 0 {
		t.Errorf("пустой фильтр взял %d профилей, ожидался пустой отбор", len(taken))
	}
}

func TestCoverageLineNamesTakenAndTotal(t *testing.T) {
	const want = "взято 0 профилей из 96"
	if got := coverageLine(0, len(profile.All())); got != want {
		t.Errorf("строка охвата %q, ожидалась %q", got, want)
	}
}

func TestSummaryLineNamesPassedAndFailed(t *testing.T) {
	const want = "прошло 1, упало 2 из 3 взятых"
	if got := summaryLine(runResult{taken: 3, passed: 1, failed: 2}); got != want {
		t.Errorf("сводка %q, ожидалась %q", got, want)
	}
}

func TestHasTargetReportsMissingMakefile(t *testing.T) {
	err := hasTarget(t.TempDir(), "build")
	if err == nil {
		t.Fatal("проверка цели вернула успех, ожидался отказ")
	}

	if !strings.Contains(err.Error(), errTargetMissing.Error()) {
		t.Errorf("причина %q не названа причиной «цель отсутствует»", err)
	}
}

func TestHasTargetFindsDeclaredTarget(t *testing.T) {
	dir := t.TempDir()
	makefile := ".PHONY: build\n\nbuild:\n\t@true\n"

	if err := os.WriteFile(filepath.Join(dir, makefileName), []byte(makefile), filePerm); err != nil {
		t.Fatalf("запись оснастки вернула ошибку: %v", err)
	}

	if err := hasTarget(dir, "build"); err != nil {
		t.Errorf("объявленная цель не найдена: %v", err)
	}

	if err := hasTarget(dir, "lint"); err == nil {
		t.Error("необъявленная цель найдена, ожидался отказ")
	}
}
