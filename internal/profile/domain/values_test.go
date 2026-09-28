package domain_test

import (
	"errors"
	"testing"

	"github.com/united-software-platform/extruder/internal/profile/domain"
)

func TestNewLanguageAcceptsAllowedValue(t *testing.T) {
	language, err := domain.NewLanguage("go")
	if err != nil {
		t.Fatalf("допустимое значение отвергнуто: %v", err)
	}
	if language.String() != "go" {
		t.Fatalf("получено %q, ожидалось %q", language.String(), "go")
	}
}

func TestNewLanguageRejectsUnknownValue(t *testing.T) {
	_, err := domain.NewLanguage("rust")
	var invalid domain.InvalidValueError
	if !errors.As(err, &invalid) {
		t.Fatalf("ожидалась InvalidValueError, получено: %v", err)
	}
	if invalid.Dimension != domain.DimensionLanguage {
		t.Errorf("измерение %q, ожидалось %q", invalid.Dimension, domain.DimensionLanguage)
	}
	if invalid.Value != "rust" {
		t.Errorf("значение %q, ожидалось %q", invalid.Value, "rust")
	}
	if len(invalid.Allowed) == 0 {
		t.Error("перечень допустимых значений пуст")
	}
}

func TestProjectTypeAndArchitectureValidateValue(t *testing.T) {
	if _, err := domain.NewProjectType("monolith"); err == nil {
		t.Error("недопустимый тип проекта принят")
	}
	if _, err := domain.NewArchitecture("clean"); err == nil {
		t.Error("недопустимая архитектура принята")
	}
}

func TestNewPatternsAllowsEmptySet(t *testing.T) {
	patterns, err := domain.NewPatterns(nil)
	if err != nil {
		t.Fatalf("пустой набор отвергнут: %v", err)
	}
	if !patterns.IsEmpty() {
		t.Error("набор без значений не считается пустым")
	}
}

func TestNewPatternsKeepsTwoValues(t *testing.T) {
	patterns, err := domain.NewPatterns([]string{"ddd", "cqrs"})
	if err != nil {
		t.Fatalf("допустимый набор отвергнут: %v", err)
	}
	values := patterns.Values()
	if len(values) != 2 || values[0] != "ddd" || values[1] != "cqrs" {
		t.Fatalf("состав %v, ожидался [ddd cqrs]", values)
	}
}

func TestNewPatternsRejectsDuplicate(t *testing.T) {
	_, err := domain.NewPatterns([]string{"ddd", "ddd"})
	var duplicate domain.DuplicateValueError
	if !errors.As(err, &duplicate) {
		t.Fatalf("ожидалась DuplicateValueError, получено: %v", err)
	}
	if duplicate.Value != "ddd" {
		t.Errorf("значение %q, ожидалось %q", duplicate.Value, "ddd")
	}
}

func TestNewPatternsRejectsUnknownValue(t *testing.T) {
	if _, err := domain.NewPatterns([]string{"event-sourcing"}); err == nil {
		t.Error("недопустимый паттерн принят")
	}
}

func TestAllowedValuesReturnsCopy(t *testing.T) {
	first := domain.AllowedValues(domain.DimensionLanguage)
	if len(first) == 0 {
		t.Fatal("перечень допустимых языков пуст")
	}
	first[0] = "испорчено"
	second := domain.AllowedValues(domain.DimensionLanguage)
	if second[0] == "испорчено" {
		t.Error("правка возвращённого перечня изменила перечень домена")
	}
}
