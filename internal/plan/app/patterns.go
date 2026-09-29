package app

import (
	profile "github.com/united-software-platform/extruder/internal/profile/domain"
)

// patternsContribution вносит вклад измерения «подходы и паттерны»: дополнительные роли
// в существующие слои. Пустой набор паттернов вклада не даёт.
//
// Роли паттернов появляются на этапе 7 карты плана — change `add-pattern-dimension`: перечень
// ролей этой версии их ещё не содержит. Вкладчик существует как точка расширения и проверяет,
// что каждое допущенное профилем значение плану известно.
func patternsContribution(patterns profile.Patterns) ([]layerRole, error) {
	var added []layerRole
	for _, pattern := range patterns.Values() {
		switch pattern {
		case "ddd", "cqrs":
			continue
		default:
			return nil, UnsupportedValueError{
				Dimension: profile.DimensionPatterns,
				Value:     pattern,
			}
		}
	}
	return added, nil
}
