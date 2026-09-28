// Package domain содержит модель архитектурного профиля: измерения, профиль и правила,
// по которым значение измерения признаётся допустимым. Слой о командной строке не знает.
package domain

import (
	"fmt"
	"strings"
)

// Dimension — независимая ось выбора в профиле. Значение типа используется в сообщениях
// пользователю, поэтому хранит имя измерения так, как оно называется в документации.
type Dimension string

// Измерения профиля. Перечень закрыт: новое измерение — изменение модели, а не данных.
const (
	DimensionLanguage     Dimension = "язык"
	DimensionProjectType  Dimension = "тип проекта"
	DimensionArchitecture Dimension = "архитектура"
	DimensionPatterns     Dimension = "подходы и паттерны"
)

// InvalidValueError сообщает о значении, которого нет в перечне допустимых.
// Несёт три части, которых требует сообщение об отказе: измерение, полученное
// значение и перечень допустимых значений этого измерения.
type InvalidValueError struct {
	Dimension Dimension
	Value     string
	Allowed   []string
}

func (e InvalidValueError) Error() string {
	return fmt.Sprintf(
		"измерение %q: недопустимое значение %q; допустимые значения: %s",
		string(e.Dimension), e.Value, strings.Join(e.Allowed, ", "),
	)
}

// DuplicateValueError сообщает о повторе значения там, где принимается набор.
// Повтор не сворачивается молча: он означает ошибку ввода, а не намерение.
type DuplicateValueError struct {
	Dimension Dimension
	Value     string
}

func (e DuplicateValueError) Error() string {
	return fmt.Sprintf("измерение %q: значение %q указано дважды", string(e.Dimension), e.Value)
}
