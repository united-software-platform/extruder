package infrastructure

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/united-software-platform/extruder/internal/profile/application"
)

// Prompt реализует контракт опроса слоя приложения.
var _ application.Prompter = (*Prompt)(nil)

// Причины отказа при неразобранном ответе. Печатаются перед повторным предъявлением того же
// вопроса: число попыток не ограничено, поэтому причина — единственное, что отличает повтор
// от первого предъявления.
const (
	reasonEmptyAnswer = "Ответ пуст."
	reasonNotANumber  = "Ответ — не номер пункта."
	reasonOutOfRange  = "Пункта с таким номером в списке нет."
)

// answerSeparator — разделитель номеров в ответе на вопрос с множественным выбором.
const answerSeparator = ","

// answerPrefix — приглашение к ответу, печатаемое после вопроса без перевода строки.
const answerPrefix = "Ответ: "

// Prompt — опрос поверх потоков ввода и вывода: нумерация пунктов, печать вопроса, разбор строки
// ответа и повтор вопроса при неразобранном ответе. Терминал не эмулируется, поэтому опрос
// проверяется тестом на подставленном потоке.
type Prompt struct {
	in  *bufio.Reader
	out io.Writer
}

// NewPrompt возвращает опрос, читающий ответы из in и печатающий вопросы в out.
func NewPrompt(in io.Reader, out io.Writer) *Prompt {
	return &Prompt{in: bufio.NewReader(in), out: out}
}

// SelectOne печатает вопрос нумерованным списком и читает номер пункта. Неразобранный ответ
// печатается причиной, и тот же вопрос предъявляется снова.
func (p *Prompt) SelectOne(prompt string, options []string) (int, error) {
	reason := ""

	for {
		answer, err := p.ask(reason, prompt, options)
		if err != nil {
			return 0, err
		}

		index, rejected := parseNumber(answer, len(options))
		if rejected != "" {
			reason = rejected

			continue
		}

		return index, nil
	}
}

// SelectMany печатает вопрос нумерованным списком и читает номера пунктов через запятую. Повтор
// номера в ответе значения не удваивает; порядок результата задаёт список пунктов.
func (p *Prompt) SelectMany(prompt string, options []string) ([]int, error) {
	reason := ""

	for {
		answer, err := p.ask(reason, prompt, options)
		if err != nil {
			return nil, err
		}

		indexes, rejected := parseNumbers(answer, len(options))
		if rejected != "" {
			reason = rejected

			continue
		}

		return indexes, nil
	}
}

// ReadLine печатает подсказку и возвращает прочитанную строку. Годность строки решает
// вызывающий сценарий: пустая строка возвращается как есть, и переспрашивает её он сам.
func (p *Prompt) ReadLine(prompt string) (string, error) {
	if err := p.print(prompt); err != nil {
		return "", err
	}

	return p.readLine()
}

// ask печатает причину отказа предыдущего ответа, вопрос, нумерованный список пунктов
// и приглашение, после чего читает строку ответа.
func (p *Prompt) ask(reason, prompt string, options []string) (string, error) {
	lines := make([]string, 0, len(options)+2)

	if reason != "" {
		lines = append(lines, reason)
	}

	lines = append(lines, prompt)

	for index, option := range options {
		lines = append(lines, fmt.Sprintf("  %d) %s", index+1, option))
	}

	if err := p.print(strings.Join(lines, "\n")); err != nil {
		return "", err
	}

	return p.readLine()
}

// print печатает текст и приглашение к ответу.
func (p *Prompt) print(text string) error {
	if _, err := fmt.Fprintln(p.out, text); err != nil {
		return fmt.Errorf("печать вопроса: %w", err)
	}

	if _, err := fmt.Fprint(p.out, answerPrefix); err != nil {
		return fmt.Errorf("печать приглашения к ответу: %w", err)
	}

	return nil
}

// readLine читает строку ответа без обрамляющих пробелов. Конец потока до ответа — отказ
// «ввод прерван»: это ошибка запуска, а не решение человека.
func (p *Prompt) readLine() (string, error) {
	line, err := p.in.ReadString('\n')

	switch {
	case err == nil:
		return strings.TrimSpace(line), nil
	case errors.Is(err, io.EOF) && strings.TrimSpace(line) != "":
		// Последняя строка потока без перевода строки — полноценный ответ.
		return strings.TrimSpace(line), nil
	case errors.Is(err, io.EOF):
		return "", fmt.Errorf("%w: поток ввода кончился до ответа", application.ErrInputInterrupted)
	default:
		return "", fmt.Errorf("%w: чтение ответа: %w", application.ErrInputInterrupted, err)
	}
}

// parseNumber разбирает ответ как номер одного пункта и возвращает его, считая от нуля.
// Непустая вторая величина — причина, по которой ответ не разобран.
func parseNumber(answer string, count int) (int, string) {
	if answer == "" {
		return 0, reasonEmptyAnswer
	}

	number, err := strconv.Atoi(answer)
	if err != nil {
		return 0, reasonNotANumber
	}

	if number < 1 || number > count {
		return 0, reasonOutOfRange
	}

	return number - 1, ""
}

// parseNumbers разбирает ответ как перечень номеров через запятую: повторы снимаются, порядок
// результата — порядок пунктов списка.
func parseNumbers(answer string, count int) ([]int, string) {
	if answer == "" {
		return nil, reasonEmptyAnswer
	}

	indexes := make([]int, 0, count)

	for _, part := range strings.Split(answer, answerSeparator) {
		index, rejected := parseNumber(strings.TrimSpace(part), count)
		if rejected != "" {
			return nil, rejected
		}

		if !slices.Contains(indexes, index) {
			indexes = append(indexes, index)
		}
	}

	slices.Sort(indexes)

	return indexes, ""
}
