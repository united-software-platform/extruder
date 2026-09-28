package domain_test

import (
	"testing"

	"github.com/united-software-platform/extruder/internal/profile/domain"
)

func buildProfile(t *testing.T, patterns ...string) domain.Profile {
	t.Helper()
	language, err := domain.NewLanguage("go")
	if err != nil {
		t.Fatalf("язык: %v", err)
	}
	projectType, err := domain.NewProjectType("service")
	if err != nil {
		t.Fatalf("тип проекта: %v", err)
	}
	architecture, err := domain.NewArchitecture("layered")
	if err != nil {
		t.Fatalf("архитектура: %v", err)
	}
	set, err := domain.NewPatterns(patterns)
	if err != nil {
		t.Fatalf("паттерны: %v", err)
	}
	return domain.NewProfile(language, projectType, architecture, set)
}

func TestProfileEqualForSameValues(t *testing.T) {
	first := buildProfile(t, "ddd")
	second := buildProfile(t, "ddd")
	if !first.Equal(second) {
		t.Error("профили с одинаковыми значениями не равны")
	}
}

func TestProfileNotEqualForDifferentPatterns(t *testing.T) {
	first := buildProfile(t, "ddd")
	second := buildProfile(t)
	if first.Equal(second) {
		t.Error("профили с разным составом паттернов признаны равными")
	}
}
