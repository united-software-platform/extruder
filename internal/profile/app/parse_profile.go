// Package app содержит сценарии приложения модуля «профиль»: порядок шагов операции,
// но не бизнес-правила. Правила проверки значений живут в домене.
package app

import (
	"strings"

	"github.com/united-software-platform/extruder/internal/profile/domain"
)

// ParseProfileInput — входной контракт сценария: сырые значения измерений, как они пришли
// извне. Неизменяем и логики не содержит.
type ParseProfileInput struct {
	Language     string
	ProjectType  string
	Architecture string
	Patterns     []string
}

// ParseProfileOutput — выходной контракт сценария: разобранный профиль.
type ParseProfileOutput struct {
	Profile domain.Profile
}

// MissingDimensionsError сообщает о незаполненных обязательных измерениях. Несёт все
// недостающие измерения сразу: пользователю нужен полный перечень, а не первое имя.
type MissingDimensionsError struct {
	Dimensions []domain.Dimension
}

func (e MissingDimensionsError) Error() string {
	names := make([]string, 0, len(e.Dimensions))
	for _, dimension := range e.Dimensions {
		names = append(names, "\""+string(dimension)+"\"")
	}
	return "не указаны обязательные измерения: " + strings.Join(names, ", ")
}

// ValidationError собирает все обнаруженные ошибки ввода. Разбор не останавливается
// на первой: пользователь должен увидеть весь список, а не исправлять ошибки по одной.
type ValidationError struct {
	Errors []error
}

func (e ValidationError) Error() string {
	messages := make([]string, 0, len(e.Errors))
	for _, err := range e.Errors {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "\n")
}

// Unwrap открывает вложенные ошибки для errors.As и errors.Is.
func (e ValidationError) Unwrap() []error { return e.Errors }

// ParseProfileUseCase — сценарий разбора профиля из сырых значений измерений.
type ParseProfileUseCase struct{}

// NewParseProfileUseCase создаёт сценарий.
func NewParseProfileUseCase() ParseProfileUseCase { return ParseProfileUseCase{} }

// Execute разбирает профиль. Единственная публичная операция сценария: принимает входной
// контракт и возвращает выходной либо перечень обнаруженных ошибок ввода.
func (uc ParseProfileUseCase) Execute(input ParseProfileInput) (ParseProfileOutput, error) {
	var problems []error

	if missing := collectMissing(input); len(missing) > 0 {
		problems = append(problems, MissingDimensionsError{Dimensions: missing})
	}

	language, err := createIfPresent(input.Language, domain.NewLanguage)
	if err != nil {
		problems = append(problems, err)
	}
	projectType, err := createIfPresent(input.ProjectType, domain.NewProjectType)
	if err != nil {
		problems = append(problems, err)
	}
	architecture, err := createIfPresent(input.Architecture, domain.NewArchitecture)
	if err != nil {
		problems = append(problems, err)
	}
	patterns, err := domain.NewPatterns(input.Patterns)
	if err != nil {
		problems = append(problems, err)
	}

	if len(problems) > 0 {
		return ParseProfileOutput{}, ValidationError{Errors: problems}
	}
	return ParseProfileOutput{
		Profile: domain.NewProfile(language, projectType, architecture, patterns),
	}, nil
}

// collectMissing возвращает обязательные измерения, для которых значение не задано.
func collectMissing(input ParseProfileInput) []domain.Dimension {
	var missing []domain.Dimension
	if input.Language == "" {
		missing = append(missing, domain.DimensionLanguage)
	}
	if input.ProjectType == "" {
		missing = append(missing, domain.DimensionProjectType)
	}
	if input.Architecture == "" {
		missing = append(missing, domain.DimensionArchitecture)
	}
	return missing
}

// createIfPresent вызывает конструктор измерения только для непустого значения:
// о пустом значении уже сообщила проверка обязательности, и второе сообщение о том же
// измерении было бы шумом.
func createIfPresent[T any](value string, create func(string) (T, error)) (T, error) {
	if value == "" {
		var zero T
		return zero, nil
	}
	return create(value)
}
