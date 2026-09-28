package app_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/united-software-platform/extruder/internal/profile/app"
	"github.com/united-software-platform/extruder/internal/profile/domain"
)

func fullInput() app.ParseProfileInput {
	return app.ParseProfileInput{
		Language:     "go",
		ProjectType:  "service",
		Architecture: "layered",
		Patterns:     []string{"ddd"},
	}
}

func TestExecuteParsesFullProfile(t *testing.T) {
	output, err := app.NewParseProfileUseCase().Execute(fullInput())
	if err != nil {
		t.Fatalf("полный профиль отвергнут: %v", err)
	}
	if output.Profile.Language().String() != "go" {
		t.Errorf("язык %q, ожидался %q", output.Profile.Language(), "go")
	}
	if output.Profile.Patterns().IsEmpty() {
		t.Error("паттерны потеряны")
	}
}

func TestExecuteAcceptsProfileWithoutPatterns(t *testing.T) {
	input := fullInput()
	input.Patterns = nil
	output, err := app.NewParseProfileUseCase().Execute(input)
	if err != nil {
		t.Fatalf("профиль без паттернов отвергнут: %v", err)
	}
	if !output.Profile.Patterns().IsEmpty() {
		t.Error("набор паттернов не пуст")
	}
}

func TestExecuteReportsSingleMissingDimension(t *testing.T) {
	input := fullInput()
	input.Architecture = ""
	_, err := app.NewParseProfileUseCase().Execute(input)
	if err == nil {
		t.Fatal("профиль без архитектуры принят")
	}
	if !strings.Contains(err.Error(), string(domain.DimensionArchitecture)) {
		t.Errorf("сообщение не называет измерение «архитектура»: %s", err)
	}
}

func TestExecuteReportsEveryMissingDimension(t *testing.T) {
	input := fullInput()
	input.ProjectType = ""
	input.Architecture = ""
	_, err := app.NewParseProfileUseCase().Execute(input)
	if err == nil {
		t.Fatal("профиль без двух измерений принят")
	}
	var missing app.MissingDimensionsError
	if !errors.As(err, &missing) {
		t.Fatalf("ожидалась MissingDimensionsError, получено: %v", err)
	}
	if len(missing.Dimensions) != 2 {
		t.Fatalf("названо %d измерений, ожидалось 2: %v", len(missing.Dimensions), missing.Dimensions)
	}
	message := err.Error()
	for _, dimension := range []domain.Dimension{domain.DimensionProjectType, domain.DimensionArchitecture} {
		if !strings.Contains(message, string(dimension)) {
			t.Errorf("сообщение не называет измерение %q: %s", string(dimension), message)
		}
	}
}

func TestExecuteReportsInvalidValueWithAllowedList(t *testing.T) {
	input := fullInput()
	input.Language = "rust"
	_, err := app.NewParseProfileUseCase().Execute(input)
	if err == nil {
		t.Fatal("неизвестный язык принят")
	}
	var invalid domain.InvalidValueError
	if !errors.As(err, &invalid) {
		t.Fatalf("ожидалась InvalidValueError, получено: %v", err)
	}
	message := err.Error()
	for _, part := range []string{string(domain.DimensionLanguage), "rust", "go"} {
		if !strings.Contains(message, part) {
			t.Errorf("сообщение не содержит %q: %s", part, message)
		}
	}
}

func TestExecuteDoesNotSubstituteNearestAllowedValue(t *testing.T) {
	input := fullInput()
	input.Language = "golang"
	output, err := app.NewParseProfileUseCase().Execute(input)
	if err == nil {
		t.Fatal("недопустимое значение принято")
	}
	if output.Profile.Language().String() != "" {
		t.Errorf("значение подставлено вместо отказа: %q", output.Profile.Language())
	}
}

func TestExecuteRejectsDuplicatePattern(t *testing.T) {
	input := fullInput()
	input.Patterns = []string{"ddd", "ddd"}
	_, err := app.NewParseProfileUseCase().Execute(input)
	var duplicate domain.DuplicateValueError
	if !errors.As(err, &duplicate) {
		t.Fatalf("ожидалась DuplicateValueError, получено: %v", err)
	}
}

func TestExecuteCollectsSeveralProblemsAtOnce(t *testing.T) {
	_, err := app.NewParseProfileUseCase().Execute(app.ParseProfileInput{
		ProjectType: "monolith",
	})
	var validation app.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("ожидалась ValidationError, получено: %v", err)
	}
	if len(validation.Errors) < 2 {
		t.Errorf("собрано %d ошибок, ожидалось не меньше 2", len(validation.Errors))
	}
}
