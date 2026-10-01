//go:build acceptance

// Package acceptance — приёмочный прогон Extruder: порождение профилей из словаря, отбор фильтром
// реализованных значений, проверка трёх свойств сгенерированного проекта и предъявление охвата
// и вердиктов.
//
// Прогон — проверка самого Extruder, а не часть поставляемого инструмента: тег сборки `acceptance`
// выводит его из обычного прогона тестов, а цель `go-acceptance` запускает отдельно.
package acceptance

import (
	"fmt"
	"slices"

	profile "github.com/united-software-platform/extruder/internal/profile/domain"
)

// implementedValues — фильтр реализованных значений: значения измерений, накладки которых
// генератор уже умеет применять.
//
// Фильтр живёт в пакете прогона, а не в домене: готовность накладки — свойство реализации
// генератора, а не правило предметной области. Домен, знающий, что «уже сделано», начал бы
// расходиться со словарём.
type implementedValues struct {
	languages     []profile.Language
	projectTypes  []profile.ProjectType
	architectures []profile.Architecture
	patterns      []profile.Pattern
}

// implemented возвращает действующий фильтр реализованных значений. Пока генератор — заглушка,
// фильтр пуст: ни один профиль в прогон не берётся, и прогон штатно зелёный.
func implemented() implementedValues {
	return implementedValues{}
}

// allows сообщает, берётся ли профиль в прогон: реализованы должны быть его язык, тип проекта,
// архитектура и каждое значение набора паттернов. Пустой набор паттернов требований
// не предъявляет — профиль без паттернов проверяем, как только готовы три обязательных измерения.
func (f implementedValues) allows(p profile.Profile) bool {
	if !slices.Contains(f.languages, p.Language()) {
		return false
	}

	if !slices.Contains(f.projectTypes, p.ProjectType()) {
		return false
	}

	if !slices.Contains(f.architectures, p.Architecture()) {
		return false
	}

	for _, pattern := range p.Patterns() {
		if !slices.Contains(f.patterns, pattern) {
			return false
		}
	}

	return true
}

// selectProfiles отбирает фильтром профили, которые берутся в прогон. Профили приходят перечнем
// допустимых: собственной комбинаторики у прогона нет — иначе число допустимых сочетаний стало бы
// побочным результатом цикла и разошлось бы со словарём незаметно.
func selectProfiles(all []profile.Profile, filter implementedValues) []profile.Profile {
	taken := make([]profile.Profile, 0, len(all))

	for _, p := range all {
		if filter.allows(p) {
			taken = append(taken, p)
		}
	}

	return taken
}

// coverageLine — строка охвата прогона. Печатается всегда, в том числе при нулевом числе взятых
// профилей: по ней читается охват проверки на текущем этапе работ, и без неё пустой фильтр
// неотличим от зелёной проверки всех сочетаний.
func coverageLine(taken, total int) string {
	return fmt.Sprintf("взято %d профилей из %d", taken, total)
}

// summaryLine — итоговая сводка прогона: числа прошедших и упавших профилей из взятых.
// Сводкой опроса она не является — это другой контекст, проверка самого Extruder.
func summaryLine(result runResult) string {
	return fmt.Sprintf(
		"прошло %d, упало %d из %d взятых",
		result.passed, result.failed, result.taken,
	)
}

// projectIdentifier — идентификатор проекта, передаваемый генератору. Прогон не знает языка
// профиля, поэтому берёт строку, допустимую и как путь go-модуля, и как имя дистрибутива Python.
func projectIdentifier(p profile.Profile) string {
	return "acceptance-" + p.String()
}
