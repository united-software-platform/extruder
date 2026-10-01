package infrastructure

import (
	"bytes"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/united-software-platform/extruder/internal/profile/application"
)

// dimensionOptions — пункты вопроса об измерении, которыми проверяется опрос.
func dimensionOptions() []string {
	return []string{"ddd", "cqrs", "без паттернов"}
}

// newTestPrompt собирает опрос на подставленном потоке ответов.
func newTestPrompt(answers string) (*Prompt, *bytes.Buffer) {
	var out bytes.Buffer

	return NewPrompt(strings.NewReader(answers), &out), &out
}

func TestSelectOnePrintsNumberedList(t *testing.T) {
	prompt, out := newTestPrompt("2\n")

	index, err := prompt.SelectOne("Выберите паттерн", dimensionOptions())
	if err != nil {
		t.Fatalf("выбор пункта вернул ошибку: %v", err)
	}

	if index != 1 {
		t.Errorf("выбран пункт %d, ожидался %d", index, 1)
	}

	printed := out.String()
	for _, line := range []string{"Выберите паттерн", "  1) ddd", "  2) cqrs", "  3) без паттернов", answerPrefix} {
		if !strings.Contains(printed, line) {
			t.Errorf("в выводе нет %q:\n%s", line, printed)
		}
	}
}

func TestSelectOneRepeatsQuestionOnRejectedAnswer(t *testing.T) {
	cases := []struct {
		name    string
		answers string
		reason  string
	}{
		{name: "номер вне списка", answers: "9\n1\n", reason: reasonOutOfRange},
		{name: "не число", answers: "ddd\n1\n", reason: reasonNotANumber},
		{name: "пустой ответ", answers: "\n1\n", reason: reasonEmptyAnswer},
		{name: "ноль", answers: "0\n1\n", reason: reasonOutOfRange},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			prompt, out := newTestPrompt(test.answers)

			index, err := prompt.SelectOne("Выберите паттерн", dimensionOptions())
			if err != nil {
				t.Fatalf("выбор пункта вернул ошибку: %v", err)
			}

			if index != 0 {
				t.Errorf("выбран пункт %d, ожидался %d", index, 0)
			}

			printed := out.String()
			if !strings.Contains(printed, test.reason) {
				t.Errorf("причина отказа %q не напечатана:\n%s", test.reason, printed)
			}

			if asked := strings.Count(printed, "Выберите паттерн"); asked != 2 {
				t.Errorf("вопрос задан %d раз, ожидалось 2:\n%s", asked, printed)
			}
		})
	}
}

func TestSelectManyReadsNumbersSeparatedByComma(t *testing.T) {
	prompt, _ := newTestPrompt("2, 1\n")

	indexes, err := prompt.SelectMany("Выберите паттерны", dimensionOptions())
	if err != nil {
		t.Fatalf("выбор пунктов вернул ошибку: %v", err)
	}

	want := []int{0, 1}
	if !slices.Equal(indexes, want) {
		t.Errorf("выбраны пункты %v, ожидались %v в порядке списка", indexes, want)
	}
}

func TestSelectManyDropsRepeatedNumber(t *testing.T) {
	prompt, _ := newTestPrompt("1,1\n")

	indexes, err := prompt.SelectMany("Выберите паттерны", dimensionOptions())
	if err != nil {
		t.Fatalf("выбор пунктов вернул ошибку: %v", err)
	}

	want := []int{0}
	if !slices.Equal(indexes, want) {
		t.Errorf("выбраны пункты %v, ожидались %v: повтор номера удвоил значение", indexes, want)
	}
}

func TestSelectManySelectsNoPatternsItem(t *testing.T) {
	prompt, _ := newTestPrompt("3\n")

	indexes, err := prompt.SelectMany("Выберите паттерны", dimensionOptions())
	if err != nil {
		t.Fatalf("выбор пунктов вернул ошибку: %v", err)
	}

	want := []int{2}
	if !slices.Equal(indexes, want) {
		t.Errorf("выбраны пункты %v, ожидались %v", indexes, want)
	}
}

func TestSelectManyRepeatsQuestionOnRejectedAnswer(t *testing.T) {
	prompt, out := newTestPrompt("\n1,9\n2\n")

	indexes, err := prompt.SelectMany("Выберите паттерны", dimensionOptions())
	if err != nil {
		t.Fatalf("выбор пунктов вернул ошибку: %v", err)
	}

	if want := []int{1}; !slices.Equal(indexes, want) {
		t.Errorf("выбраны пункты %v, ожидались %v", indexes, want)
	}

	printed := out.String()
	if asked := strings.Count(printed, "Выберите паттерны"); asked != 3 {
		t.Errorf("вопрос задан %d раз, ожидалось 3:\n%s", asked, printed)
	}

	for _, reason := range []string{reasonEmptyAnswer, reasonOutOfRange} {
		if !strings.Contains(printed, reason) {
			t.Errorf("причина отказа %q не напечатана:\n%s", reason, printed)
		}
	}
}

func TestReadLineReturnsAnswerAsIs(t *testing.T) {
	prompt, out := newTestPrompt("  github.com/acme/orders  \n")

	answer, err := prompt.ReadLine("Введите идентификатор проекта")
	if err != nil {
		t.Fatalf("чтение строки вернуло ошибку: %v", err)
	}

	if answer != "github.com/acme/orders" {
		t.Errorf("прочитано %q, ожидалось %q", answer, "github.com/acme/orders")
	}

	if !strings.Contains(out.String(), "Введите идентификатор проекта") {
		t.Errorf("подсказка не напечатана:\n%s", out.String())
	}
}

func TestReadLineReturnsEmptyAnswer(t *testing.T) {
	prompt, _ := newTestPrompt("\n")

	answer, err := prompt.ReadLine("Продолжить? (да/нет)")
	if err != nil {
		t.Fatalf("чтение строки вернуло ошибку: %v", err)
	}

	if answer != "" {
		t.Errorf("прочитано %q, ожидалась пустая строка: решение о пустом ответе принимает сценарий", answer)
	}
}

func TestReadLineAcceptsLastLineWithoutNewline(t *testing.T) {
	prompt, _ := newTestPrompt("orders")

	answer, err := prompt.ReadLine("Введите идентификатор проекта")
	if err != nil {
		t.Fatalf("чтение строки вернуло ошибку: %v", err)
	}

	if answer != "orders" {
		t.Errorf("прочитано %q, ожидалось %q", answer, "orders")
	}
}

func TestInterruptedInputOnExhaustedStream(t *testing.T) {
	cases := []struct {
		name string
		read func(*Prompt) error
	}{
		{
			name: "вопрос с одиночным выбором",
			read: func(p *Prompt) error {
				_, err := p.SelectOne("Выберите паттерн", dimensionOptions())

				return err
			},
		},
		{
			name: "вопрос с множественным выбором",
			read: func(p *Prompt) error {
				_, err := p.SelectMany("Выберите паттерны", dimensionOptions())

				return err
			},
		},
		{
			name: "чтение строки",
			read: func(p *Prompt) error {
				_, err := p.ReadLine("Введите идентификатор проекта")

				return err
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			prompt, _ := newTestPrompt("")

			if err := test.read(prompt); !errors.Is(err, application.ErrInputInterrupted) {
				t.Errorf("ошибка %v не сравнима с ErrInputInterrupted", err)
			}
		})
	}
}

func TestInterruptedInputInTheMiddleOfSurvey(t *testing.T) {
	// Поток отвечает на первый вопрос и кончается на втором.
	prompt, _ := newTestPrompt("1\n")

	if _, err := prompt.SelectOne("Выберите язык", dimensionOptions()); err != nil {
		t.Fatalf("первый вопрос вернул ошибку: %v", err)
	}

	_, err := prompt.SelectOne("Выберите тип проекта", dimensionOptions())
	if !errors.Is(err, application.ErrInputInterrupted) {
		t.Errorf("ошибка %v не сравнима с ErrInputInterrupted", err)
	}
}
