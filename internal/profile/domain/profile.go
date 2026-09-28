package domain

import "slices"

// Profile — архитектурный профиль: набор значений всех измерений. Объект-значение:
// идентичности у профиля нет, два профиля с одинаковыми значениями — один и тот же профиль.
type Profile struct {
	language     Language
	projectType  ProjectType
	architecture Architecture
	patterns     Patterns
}

// NewProfile собирает профиль из уже проверенных измерений. Проверка значений живёт
// в конструкторах измерений, поэтому собранный профиль заведомо корректен.
func NewProfile(
	language Language,
	projectType ProjectType,
	architecture Architecture,
	patterns Patterns,
) Profile {
	return Profile{
		language:     language,
		projectType:  projectType,
		architecture: architecture,
		patterns:     patterns,
	}
}

// Language возвращает измерение «язык».
func (p Profile) Language() Language { return p.language }

// ProjectType возвращает измерение «тип проекта».
func (p Profile) ProjectType() ProjectType { return p.projectType }

// Architecture возвращает измерение «архитектура».
func (p Profile) Architecture() Architecture { return p.architecture }

// Patterns возвращает измерение «подходы и паттерны».
func (p Profile) Patterns() Patterns { return p.patterns }

// Equal сравнивает профили по значениям всех измерений. Метод нужен потому, что набор
// паттернов делает профиль несравнимым штатным сравнением Go.
func (p Profile) Equal(other Profile) bool {
	return p.language == other.language &&
		p.projectType == other.projectType &&
		p.architecture == other.architecture &&
		slices.Equal(p.patterns.values, other.patterns.values)
}
