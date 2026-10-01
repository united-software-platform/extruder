package application

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/united-software-platform/extruder/internal/profile/domain"
)

// fullPassPrompter собирает двойника на полный проход опроса: первые значения всех словарей,
// оба паттерна, допустимый идентификатор проекта и заданный ответ в сводке.
func fullPassPrompter(summaryAnswer string) *promptDouble {
	return &promptDouble{
		single:   []int{0, 0, 0},
		multiple: [][]int{{0, 1}},
		lines:    []string{"github.com/acme/orders", summaryAnswer},
	}
}

// assertPromptOrder проверяет, что вопросы заданы в ожидаемом порядке. Сверяется опознавательная
// часть подсказки, а не вся строка.
func assertPromptOrder(t *testing.T, prompts, want []string) {
	t.Helper()

	if len(prompts) != len(want) {
		t.Fatalf("задано вопросов %d, ожидалось %d: %v", len(prompts), len(want), prompts)
	}

	for index, marker := range want {
		if !strings.Contains(prompts[index], marker) {
			t.Errorf("вопрос %d — %q, ожидался содержащий %q", index+1, prompts[index], marker)
		}
	}
}

func TestCollectProfileAsksInModelOrder(t *testing.T) {
	prompter := fullPassPrompter("да")

	out, err := NewCollectProfile(prompter).Execute()
	if err != nil {
		t.Fatalf("сбор профиля вернул ошибку: %v", err)
	}

	assertPromptOrder(t, prompter.prompts, []string{
		questionLanguage,
		questionProjectType,
		questionArchitecture,
		questionPatterns,
		questionProjectIdentifier,
		"Сводка:",
	})

	if out.Profile.Language() != domain.LanguageGo {
		t.Errorf("язык %q, ожидался %q", out.Profile.Language(), domain.LanguageGo)
	}

	if out.Profile.ProjectType() != domain.ProjectTypeMonolith {
		t.Errorf("тип проекта %q, ожидался %q", out.Profile.ProjectType(), domain.ProjectTypeMonolith)
	}

	if out.Profile.Architecture() != domain.ArchitectureClean {
		t.Errorf("архитектура %q, ожидалась %q", out.Profile.Architecture(), domain.ArchitectureClean)
	}

	want := []domain.Pattern{domain.PatternDDD, domain.PatternCQRS}
	if patterns := out.Profile.Patterns(); !slices.Equal(patterns, want) {
		t.Errorf("набор паттернов %v, ожидался %v", patterns, want)
	}

	if out.ProjectIdentifier != "github.com/acme/orders" {
		t.Errorf("идентификатор проекта %q, ожидался %q", out.ProjectIdentifier, "github.com/acme/orders")
	}

	if !out.Confirmed {
		t.Error("ответ «да» в сводке не подтвердил профиль")
	}
}

func TestCollectProfileSummaryShowsProfileAndIdentifier(t *testing.T) {
	prompter := fullPassPrompter("да")

	if _, err := NewCollectProfile(prompter).Execute(); err != nil {
		t.Fatalf("сбор профиля вернул ошибку: %v", err)
	}

	summary := prompter.prompts[len(prompter.prompts)-1]
	for _, marker := range []string{"go", "monolith", "clean-architecture", "ddd", "cqrs", "github.com/acme/orders"} {
		if !strings.Contains(summary, marker) {
			t.Errorf("в сводке нет %q: %s", marker, summary)
		}
	}
}

func TestCollectProfileSelectsEmptyPatternSet(t *testing.T) {
	prompter := &promptDouble{
		single:   []int{1, 1, 1},
		multiple: [][]int{{len(domain.Patterns())}},
		lines:    []string{"orders", "да"},
	}

	out, err := NewCollectProfile(prompter).Execute()
	if err != nil {
		t.Fatalf("сбор профиля вернул ошибку: %v", err)
	}

	if patterns := out.Profile.Patterns(); len(patterns) != 0 {
		t.Errorf("набор паттернов %v, ожидался пустой", patterns)
	}

	assertPromptOrder(t, prompter.prompts, []string{
		questionLanguage,
		questionProjectType,
		questionArchitecture,
		questionPatterns,
		questionProjectIdentifier,
		"Сводка:",
	})
}

func TestCollectProfileRepeatsPatternsOnConflictingAnswer(t *testing.T) {
	prompter := &promptDouble{
		single:   []int{0, 0, 0},
		multiple: [][]int{{0, len(domain.Patterns())}, {1}},
		lines:    []string{"orders", "да"},
	}

	out, err := NewCollectProfile(prompter).Execute()
	if err != nil {
		t.Fatalf("сбор профиля вернул ошибку: %v", err)
	}

	assertPromptOrder(t, prompter.prompts, []string{
		questionLanguage,
		questionProjectType,
		questionArchitecture,
		questionPatterns,
		reasonPatternsConflict,
		questionProjectIdentifier,
		"Сводка:",
	})

	want := []domain.Pattern{domain.PatternCQRS}
	if patterns := out.Profile.Patterns(); !slices.Equal(patterns, want) {
		t.Errorf("набор паттернов %v, ожидался %v", patterns, want)
	}
}

func TestCollectProfileRepeatsProjectIdentifierQuestion(t *testing.T) {
	prompter := &promptDouble{
		single:   []int{0, 0, 0},
		multiple: [][]int{{0}},
		lines:    []string{"acme orders", "", "acme/orders", "да"},
	}

	out, err := NewCollectProfile(prompter).Execute()
	if err != nil {
		t.Fatalf("сбор профиля вернул ошибку: %v", err)
	}

	assertPromptOrder(t, prompter.prompts, []string{
		questionLanguage,
		questionProjectType,
		questionArchitecture,
		questionPatterns,
		questionProjectIdentifier,
		questionProjectIdentifier,
		questionProjectIdentifier,
		"Сводка:",
	})

	// Причина отказа напечатана перед повторно заданным вопросом, а ранее выбранные значения
	// заново не спрашивались: вопросы об измерениях заданы по одному разу.
	if !strings.Contains(prompter.prompts[5], domain.ErrInvalidProjectIdentifier.Error()) {
		t.Errorf("повторный вопрос %q не называет причину отказа", prompter.prompts[5])
	}

	if out.ProjectIdentifier != "acme/orders" {
		t.Errorf("идентификатор проекта %q, ожидался %q", out.ProjectIdentifier, "acme/orders")
	}

	if out.Profile.Language() != domain.LanguageGo {
		t.Errorf("язык %q, ожидался %q: ранее данный ответ потерян", out.Profile.Language(), domain.LanguageGo)
	}
}

func TestCollectProfileDeclinedInSummary(t *testing.T) {
	cases := []struct {
		name   string
		answer string
	}{
		{name: "ответ «нет»", answer: "нет"},
		{name: "пустой ответ", answer: ""},
		{name: "произвольный ответ", answer: "позже"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			prompter := fullPassPrompter(test.answer)

			out, err := NewCollectProfile(prompter).Execute()
			if err != nil {
				t.Fatalf("сбор профиля вернул ошибку: %v", err)
			}

			if out.Confirmed {
				t.Errorf("ответ %q в сводке подтвердил профиль, ожидался отказ", test.answer)
			}

			if out.Profile.Language() != domain.LanguageGo {
				t.Error("отказ в сводке потерял собранный профиль")
			}
		})
	}
}

func TestCollectProfileInterruptedInput(t *testing.T) {
	cases := []struct {
		name     string
		prompter *promptDouble
	}{
		{name: "поток кончился на первом вопросе", prompter: &promptDouble{}},
		{
			name:     "поток кончился на вопросе о паттернах",
			prompter: &promptDouble{single: []int{0, 0, 0}},
		},
		{
			name: "поток кончился на идентификаторе проекта",
			prompter: &promptDouble{
				single:   []int{0, 0, 0},
				multiple: [][]int{{0}},
			},
		},
		{
			name: "поток кончился в сводке",
			prompter: &promptDouble{
				single:   []int{0, 0, 0},
				multiple: [][]int{{0}},
				lines:    []string{"orders"},
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewCollectProfile(test.prompter).Execute()
			if !errors.Is(err, ErrInputInterrupted) {
				t.Errorf("ошибка %v не сравнима с ErrInputInterrupted", err)
			}
		})
	}
}
