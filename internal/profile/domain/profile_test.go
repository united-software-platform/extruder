package domain

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

// allProfilesCount — число допустимых профилей по правилу «допустимы все сочетания»:
// 2 языка × 4 типа проекта × 3 архитектуры × 4 набора паттернов.
const allProfilesCount = 96

func TestNewProfileCreatesProfileWithoutPatterns(t *testing.T) {
	got, err := NewProfile(LanguageGo, ProjectTypeMonolith, ArchitectureClean, nil)
	if err != nil {
		t.Fatalf("создание профиля вернуло ошибку: %v", err)
	}

	if got.Language() != LanguageGo {
		t.Errorf("язык %q, ожидался %q", got.Language(), LanguageGo)
	}

	if got.ProjectType() != ProjectTypeMonolith {
		t.Errorf("тип проекта %q, ожидался %q", got.ProjectType(), ProjectTypeMonolith)
	}

	if got.Architecture() != ArchitectureClean {
		t.Errorf("архитектура %q, ожидалась %q", got.Architecture(), ArchitectureClean)
	}

	if patterns := got.Patterns(); len(patterns) != 0 {
		t.Errorf("набор паттернов %v, ожидался пустой", patterns)
	}
}

func TestNewProfileCreatesProfileWithBothPatterns(t *testing.T) {
	got, err := NewProfile(
		LanguagePython,
		ProjectTypeDistributed,
		ArchitectureLayered,
		[]Pattern{PatternDDD, PatternCQRS},
	)
	if err != nil {
		t.Fatalf("создание профиля вернуло ошибку: %v", err)
	}

	want := []Pattern{PatternDDD, PatternCQRS}
	if patterns := got.Patterns(); !slices.Equal(patterns, want) {
		t.Errorf("набор паттернов %v, ожидался %v", patterns, want)
	}
}

func TestNewProfileRejectsUnknownValue(t *testing.T) {
	cases := []struct {
		name         string
		language     Language
		projectType  ProjectType
		architecture Architecture
		patterns     []Pattern
		dimension    string
		value        string
	}{
		{
			name:         "неизвестный язык",
			language:     "rust",
			projectType:  ProjectTypeMonolith,
			architecture: ArchitectureClean,
			dimension:    dimensionLanguage,
			value:        "rust",
		},
		{
			name:         "неизвестный тип проекта",
			language:     LanguageGo,
			projectType:  "microservices",
			architecture: ArchitectureClean,
			dimension:    dimensionProjectType,
			value:        "microservices",
		},
		{
			name:         "неизвестная архитектура",
			language:     LanguageGo,
			projectType:  ProjectTypeMonolith,
			architecture: "hexagonal",
			dimension:    dimensionArchitecture,
			value:        "hexagonal",
		},
		{
			name:         "неизвестное значение в наборе паттернов",
			language:     LanguageGo,
			projectType:  ProjectTypeMonolith,
			architecture: ArchitectureClean,
			patterns:     []Pattern{PatternDDD, "event-sourcing"},
			dimension:    dimensionPatterns,
			value:        "event-sourcing",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := NewProfile(
				testCase.language,
				testCase.projectType,
				testCase.architecture,
				testCase.patterns,
			)
			if err == nil {
				t.Fatalf("профиль %s создан, ожидался отказ", got)
			}

			if !errors.Is(err, ErrUnknownValue) {
				t.Errorf("ошибка %v не сравнима с ErrUnknownValue", err)
			}

			if message := err.Error(); !strings.Contains(message, testCase.dimension) {
				t.Errorf("сообщение %q не называет измерение %q", message, testCase.dimension)
			}

			if message := err.Error(); !strings.Contains(message, testCase.value) {
				t.Errorf("сообщение %q не называет отвергнутое значение %q", message, testCase.value)
			}

			if got != (Profile{}) {
				t.Errorf("при отказе возвращён профиль %s, ожидался нулевой", got)
			}
		})
	}
}

func TestNewProfileRemovesDuplicatePatterns(t *testing.T) {
	got, err := NewProfile(
		LanguageGo,
		ProjectTypeService,
		ArchitecturePortsAdapters,
		[]Pattern{PatternCQRS, PatternCQRS},
	)
	if err != nil {
		t.Fatalf("создание профиля вернуло ошибку: %v", err)
	}

	want := []Pattern{PatternCQRS}
	if patterns := got.Patterns(); !slices.Equal(patterns, want) {
		t.Errorf("набор паттернов %v, ожидался %v", patterns, want)
	}
}

func TestNewProfileNormalizesPatternOrder(t *testing.T) {
	direct, err := NewProfile(
		LanguageGo,
		ProjectTypeModularMonolith,
		ArchitectureClean,
		[]Pattern{PatternDDD, PatternCQRS},
	)
	if err != nil {
		t.Fatalf("создание первого профиля вернуло ошибку: %v", err)
	}

	reversed, err := NewProfile(
		LanguageGo,
		ProjectTypeModularMonolith,
		ArchitectureClean,
		[]Pattern{PatternCQRS, PatternDDD, PatternCQRS},
	)
	if err != nil {
		t.Fatalf("создание второго профиля вернуло ошибку: %v", err)
	}

	if direct != reversed {
		t.Errorf("профили %s и %s не равны, ожидалось равенство", direct, reversed)
	}

	want := []Pattern{PatternDDD, PatternCQRS}
	if patterns := reversed.Patterns(); !slices.Equal(patterns, want) {
		t.Errorf("набор паттернов %v, ожидался порядок применения %v", patterns, want)
	}
}

func TestAllCountsNinetySix(t *testing.T) {
	if got := len(All()); got != allProfilesCount {
		t.Errorf("в перечне допустимых профилей %d элементов, ожидалось %d", got, allProfilesCount)
	}
}

func TestAllHasNoDuplicates(t *testing.T) {
	seen := make(map[Profile]bool, allProfilesCount)

	for _, p := range All() {
		if seen[p] {
			t.Errorf("профиль %s встречается в перечне более одного раза", p)
		}

		seen[p] = true
	}
}

func TestAllContainsEveryProfileWithoutPatterns(t *testing.T) {
	seen := make(map[Profile]bool, allProfilesCount)
	for _, p := range All() {
		seen[p] = true
	}

	for _, language := range Languages() {
		for _, projectType := range ProjectTypes() {
			for _, architecture := range Architectures() {
				want, err := NewProfile(language, projectType, architecture, nil)
				if err != nil {
					t.Fatalf("создание профиля вернуло ошибку: %v", err)
				}

				if !seen[want] {
					t.Errorf("профиля %s нет в перечне допустимых профилей", want)
				}
			}
		}
	}
}

func TestAllValuesComeFromDictionaries(t *testing.T) {
	for _, p := range All() {
		if !slices.Contains(Languages(), p.Language()) {
			t.Errorf("профиль %s: язык %q вне словаря", p, p.Language())
		}

		if !slices.Contains(ProjectTypes(), p.ProjectType()) {
			t.Errorf("профиль %s: тип проекта %q вне словаря", p, p.ProjectType())
		}

		if !slices.Contains(Architectures(), p.Architecture()) {
			t.Errorf("профиль %s: архитектура %q вне словаря", p, p.Architecture())
		}

		for _, pattern := range p.Patterns() {
			if !slices.Contains(Patterns(), pattern) {
				t.Errorf("профиль %s: паттерн %q вне словаря", p, pattern)
			}
		}
	}
}

func TestProfileStringKeepsDimensionValuesIntact(t *testing.T) {
	p, err := NewProfile(
		LanguageGo,
		ProjectTypeModularMonolith,
		ArchitectureClean,
		[]Pattern{PatternCQRS, PatternDDD},
	)
	if err != nil {
		t.Fatalf("создание профиля вернуло ошибку: %v", err)
	}

	const want = "go_modular-monolith_clean-architecture_ddd-cqrs"
	if got := p.String(); got != want {
		t.Errorf("каноническое имя профиля %q, ожидалось %q", got, want)
	}
}
