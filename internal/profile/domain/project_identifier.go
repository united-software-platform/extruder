package domain

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// pathSeparator — разделитель сегментов идентификатора проекта. Ведущий и замыкающий разделитель
// запрещены: идентификатор — не путь файловой системы, а имя проекта в системе сборки.
const pathSeparator = "/"

// ErrInvalidProjectIdentifier — идентификатор проекта не прошёл общую проверку. Соответствие
// правилам системы сборки выбранного языка этой ошибкой не проверяется: правила языка приходят
// с раскладками.
var ErrInvalidProjectIdentifier = errors.New("недопустимый идентификатор проекта")

// ValidateProjectIdentifier проверяет идентификатор проекта общим правилом: непустая строка без
// пробельных и управляющих символов, без ведущего и замыкающего `/`. Отказ называет причину —
// её печатает опрос, переспрашивая тот же вопрос.
func ValidateProjectIdentifier(value string) error {
	if value == "" {
		return fmt.Errorf("%w: пустая строка", ErrInvalidProjectIdentifier)
	}

	for _, symbol := range value {
		if unicode.IsSpace(symbol) {
			return fmt.Errorf("%w: пробельный символ в строке %q", ErrInvalidProjectIdentifier, value)
		}

		if unicode.IsControl(symbol) {
			return fmt.Errorf("%w: управляющий символ в строке %q", ErrInvalidProjectIdentifier, value)
		}
	}

	if strings.HasPrefix(value, pathSeparator) {
		return fmt.Errorf("%w: ведущий %q в строке %q", ErrInvalidProjectIdentifier, pathSeparator, value)
	}

	if strings.HasSuffix(value, pathSeparator) {
		return fmt.Errorf("%w: замыкающий %q в строке %q", ErrInvalidProjectIdentifier, pathSeparator, value)
	}

	return nil
}
