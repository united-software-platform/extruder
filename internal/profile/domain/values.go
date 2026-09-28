package domain

import "slices"

// Перечни допустимых значений измерений — единственная точка определения. Отсюда их берут
// и проверка значения, и справка: разойтись они не могут, потому что источник один.
//
// Состав минимален намеренно: поддерживается ровно то, из чего собирается первый вертикальный
// срез. Объявить значение допустимым раньше, чем оно генерируется, значит пообещать
// пользователю результат, которого нет.
var (
	allowedLanguages     = []string{"go"}
	allowedProjectTypes  = []string{"service"}
	allowedArchitectures = []string{"layered"}
	allowedPatterns      = []string{"ddd", "cqrs"}
)

// AllowedValues возвращает перечень допустимых значений измерения. Копия защищает перечень
// от правки вызывающей стороной: перечень — правило домена, а не изменяемые данные.
func AllowedValues(d Dimension) []string {
	switch d {
	case DimensionLanguage:
		return slices.Clone(allowedLanguages)
	case DimensionProjectType:
		return slices.Clone(allowedProjectTypes)
	case DimensionArchitecture:
		return slices.Clone(allowedArchitectures)
	case DimensionPatterns:
		return slices.Clone(allowedPatterns)
	default:
		return nil
	}
}

// Language — измерение «язык».
type Language struct{ value string }

// NewLanguage создаёт измерение, проверяя значение при создании: недопустимое значение
// объекта не порождает, поэтому следующие шаги вход не перепроверяют.
func NewLanguage(value string) (Language, error) {
	if !slices.Contains(allowedLanguages, value) {
		return Language{}, InvalidValueError{
			Dimension: DimensionLanguage,
			Value:     value,
			Allowed:   slices.Clone(allowedLanguages),
		}
	}
	return Language{value: value}, nil
}

// String возвращает значение измерения.
func (l Language) String() string { return l.value }

// ProjectType — измерение «тип проекта».
type ProjectType struct{ value string }

// NewProjectType создаёт измерение с проверкой значения при создании.
func NewProjectType(value string) (ProjectType, error) {
	if !slices.Contains(allowedProjectTypes, value) {
		return ProjectType{}, InvalidValueError{
			Dimension: DimensionProjectType,
			Value:     value,
			Allowed:   slices.Clone(allowedProjectTypes),
		}
	}
	return ProjectType{value: value}, nil
}

// String возвращает значение измерения.
func (p ProjectType) String() string { return p.value }

// Architecture — измерение «архитектура».
type Architecture struct{ value string }

// NewArchitecture создаёт измерение с проверкой значения при создании.
func NewArchitecture(value string) (Architecture, error) {
	if !slices.Contains(allowedArchitectures, value) {
		return Architecture{}, InvalidValueError{
			Dimension: DimensionArchitecture,
			Value:     value,
			Allowed:   slices.Clone(allowedArchitectures),
		}
	}
	return Architecture{value: value}, nil
}

// String возвращает значение измерения.
func (a Architecture) String() string { return a.value }

// Patterns — измерение «подходы и паттерны». В отличие от остальных измерений принимает
// набор значений и допускает пустоту: профиль без паттернов — обычный профиль.
type Patterns struct{ values []string }

// NewPatterns создаёт измерение из набора значений. Пустой набор допустим, повтор — нет:
// повторённое значение означает ошибку ввода, а не намерение указать его дважды.
func NewPatterns(values []string) (Patterns, error) {
	seen := make(map[string]struct{}, len(values))
	kept := make([]string, 0, len(values))
	for _, value := range values {
		if !slices.Contains(allowedPatterns, value) {
			return Patterns{}, InvalidValueError{
				Dimension: DimensionPatterns,
				Value:     value,
				Allowed:   slices.Clone(allowedPatterns),
			}
		}
		if _, duplicate := seen[value]; duplicate {
			return Patterns{}, DuplicateValueError{Dimension: DimensionPatterns, Value: value}
		}
		seen[value] = struct{}{}
		kept = append(kept, value)
	}
	return Patterns{values: kept}, nil
}

// Values возвращает состав набора в порядке, в котором значения были указаны.
func (p Patterns) Values() []string { return slices.Clone(p.values) }

// IsEmpty сообщает, что паттерны не заданы.
func (p Patterns) IsEmpty() bool { return len(p.values) == 0 }
