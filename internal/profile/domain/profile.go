package domain

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Имена измерений профиля в сообщениях об отказе. Совпадают с каноническими именами единого
// языка: отказ называет измерение теми же словами, которыми его называет опрос и документы.
const (
	dimensionLanguage     = "язык"
	dimensionProjectType  = "тип проекта"
	dimensionArchitecture = "архитектура"
	dimensionPatterns     = "паттерны"
)

// Разделители канонического имени профиля. Значения измерений разделяет `_`, набор паттернов
// собирается через `-`: внутри значений измерений `-` встречается (`clean-architecture`),
// но ни одно значение паттерна его не содержит, поэтому разбор однозначен.
const (
	dimensionSeparator = "_"
	patternSeparator   = "-"
)

// ErrUnknownValue — значение, которого нет в словаре своего измерения. Профиль с таким значением
// не создаётся: проверка выполняется при создании, а не при использовании.
var ErrUnknownValue = errors.New("значение вне словаря измерения")

// Language — измерение «язык»: язык, на котором пишется начальная структура проекта.
type Language string

// Допустимые значения измерения «язык».
const (
	LanguageGo     Language = "go"
	LanguagePython Language = "python"
)

// ProjectType — измерение «тип проекта»: во сколько единиц развёртывания и модулей
// раскладывается начальная структура проекта.
type ProjectType string

// Допустимые значения измерения «тип проекта».
const (
	ProjectTypeMonolith        ProjectType = "monolith"
	ProjectTypeModularMonolith ProjectType = "modular-monolith"
	ProjectTypeService         ProjectType = "service"
	ProjectTypeDistributed     ProjectType = "distributed"
)

// Architecture — измерение «архитектура»: как внутри модуля расположены слои и куда направлены
// зависимости между ними.
type Architecture string

// Допустимые значения измерения «архитектура».
const (
	ArchitectureClean         Architecture = "clean-architecture"
	ArchitecturePortsAdapters Architecture = "ports-adapters"
	ArchitectureLayered       Architecture = "layered"
)

// Pattern — значение измерения «паттерны»: единственного измерения с множественным выбором,
// допускающего пустой набор.
type Pattern string

// Допустимые значения измерения «паттерны».
const (
	PatternDDD  Pattern = "ddd"
	PatternCQRS Pattern = "cqrs"
)

// Languages возвращает словарь измерения «язык».
func Languages() []Language {
	return []Language{LanguageGo, LanguagePython}
}

// ProjectTypes возвращает словарь измерения «тип проекта».
func ProjectTypes() []ProjectType {
	return []ProjectType{
		ProjectTypeMonolith,
		ProjectTypeModularMonolith,
		ProjectTypeService,
		ProjectTypeDistributed,
	}
}

// Architectures возвращает словарь измерения «архитектура».
func Architectures() []Architecture {
	return []Architecture{ArchitectureClean, ArchitecturePortsAdapters, ArchitectureLayered}
}

// Patterns возвращает словарь измерения «паттерны» в порядке применения: `ddd`, затем `cqrs`.
func Patterns() []Pattern {
	return []Pattern{PatternDDD, PatternCQRS}
}

// Profile — архитектурный профиль: сочетание значений четырёх независимых измерений, полностью
// определяющее начальную структуру проекта.
//
// Объект-значение: поля неэкспортируемые и после создания не меняются, инварианты проверены
// конструктором NewProfile один раз. Набор паттернов хранится канонической строкой, поэтому
// профили сравнимы оператором `==`, и порядок перечисления паттернов на равенство не влияет.
type Profile struct {
	language     Language
	projectType  ProjectType
	architecture Architecture
	patterns     string
}

// NewProfile создаёт профиль из значений четырёх измерений. Значение вне словаря своего измерения
// отклоняется ошибкой, называющей измерение и отвергнутое значение; набор паттернов нормализуется:
// повторы снимаются, порядок приводится к порядку применения.
func NewProfile(
	language Language,
	projectType ProjectType,
	architecture Architecture,
	patterns []Pattern,
) (Profile, error) {
	if !slices.Contains(Languages(), language) {
		return Profile{}, unknownValue(dimensionLanguage, string(language))
	}

	if !slices.Contains(ProjectTypes(), projectType) {
		return Profile{}, unknownValue(dimensionProjectType, string(projectType))
	}

	if !slices.Contains(Architectures(), architecture) {
		return Profile{}, unknownValue(dimensionArchitecture, string(architecture))
	}

	normalized, err := normalizePatterns(patterns)
	if err != nil {
		return Profile{}, err
	}

	return Profile{
		language:     language,
		projectType:  projectType,
		architecture: architecture,
		patterns:     normalized,
	}, nil
}

// All возвращает перечень всех допустимых профилей: измерения независимы, поэтому допустимо
// каждое сочетание их значений — 2 × 4 × 3 × 4 = 96. Перечень собирается обходом словарей,
// поэтому пополнение словаря расширяет его само, без правки где-либо ещё.
func All() []Profile {
	subsets := patternSubsets()
	all := make([]Profile, 0, len(Languages())*len(ProjectTypes())*len(Architectures())*len(subsets))

	for _, language := range Languages() {
		for _, projectType := range ProjectTypes() {
			for _, architecture := range Architectures() {
				for _, patterns := range subsets {
					all = append(all, Profile{
						language:     language,
						projectType:  projectType,
						architecture: architecture,
						patterns:     patterns,
					})
				}
			}
		}
	}

	return all
}

// Language возвращает значение измерения «язык».
func (p Profile) Language() Language {
	return p.language
}

// ProjectType возвращает значение измерения «тип проекта».
func (p Profile) ProjectType() ProjectType {
	return p.projectType
}

// Architecture возвращает значение измерения «архитектура».
func (p Profile) Architecture() Architecture {
	return p.architecture
}

// Patterns возвращает нормализованный набор паттернов: без повторов, в порядке применения.
// Пустому набору соответствует пустой результат.
func (p Profile) Patterns() []Pattern {
	if p.patterns == "" {
		return nil
	}

	values := strings.Split(p.patterns, patternSeparator)
	patterns := make([]Pattern, 0, len(values))

	for _, value := range values {
		patterns = append(patterns, Pattern(value))
	}

	return patterns
}

// String возвращает каноническое имя профиля: значения измерений через `_`, набор паттернов —
// последним сегментом через `-`. Профиль без паттернов четвёртого сегмента не получает.
// Имя пригодно и для вердикта прогона, и для имени каталога, и для идентификатора проекта.
func (p Profile) String() string {
	name := strings.Join(
		[]string{string(p.language), string(p.projectType), string(p.architecture)},
		dimensionSeparator,
	)

	if p.patterns == "" {
		return name
	}

	return name + dimensionSeparator + p.patterns
}

// unknownValue собирает отказ, называющий измерение и отвергнутое значение.
func unknownValue(dimension, value string) error {
	return fmt.Errorf("%w: измерение %q, значение %q", ErrUnknownValue, dimension, value)
}

// normalizePatterns проверяет значения набора паттернов и приводит набор к канонической строке:
// повторы сняты, порядок — порядок применения. Без нормализации наборы `{cqrs, ddd}` и
// `{ddd, cqrs}` дали бы два разных профиля, и число 96 перестало бы сходиться.
func normalizePatterns(patterns []Pattern) (string, error) {
	dictionary := Patterns()
	selected := make(map[Pattern]bool, len(patterns))

	for _, pattern := range patterns {
		if !slices.Contains(dictionary, pattern) {
			return "", unknownValue(dimensionPatterns, string(pattern))
		}

		selected[pattern] = true
	}

	return join(dictionary, selected), nil
}

// patternSubsets возвращает все подмножества словаря паттернов канонической строкой — от пустого
// набора до полного. Их число — 2 в степени размера словаря, и именно оно даёт в перечне профилей
// четвёртый множитель.
func patternSubsets() []string {
	dictionary := Patterns()
	subsets := make([]string, 0, 1<<len(dictionary))

	for mask := 0; mask < 1<<len(dictionary); mask++ {
		selected := make(map[Pattern]bool, len(dictionary))

		for index, pattern := range dictionary {
			if mask&(1<<index) != 0 {
				selected[pattern] = true
			}
		}

		subsets = append(subsets, join(dictionary, selected))
	}

	return subsets
}

// join собирает отобранные паттерны в каноническую строку, обходя словарь: порядок результата
// задаёт словарь, а не порядок, в котором значения перечислил вызывающий.
func join(dictionary []Pattern, selected map[Pattern]bool) string {
	ordered := make([]string, 0, len(dictionary))

	for _, pattern := range dictionary {
		if selected[pattern] {
			ordered = append(ordered, string(pattern))
		}
	}

	return strings.Join(ordered, patternSeparator)
}
