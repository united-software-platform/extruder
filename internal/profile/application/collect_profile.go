package application

import (
	"fmt"
	"slices"
	"strings"

	"github.com/united-software-platform/extruder/internal/profile/domain"
)

// Тексты вопросов опроса. Порядок, в котором они задаются, повторяет порядок сборки модели
// проекта и не зависит от выбранных значений.
const (
	questionLanguage          = "Выберите язык"
	questionProjectType       = "Выберите тип проекта"
	questionArchitecture      = "Выберите архитектуру"
	questionPatterns          = "Выберите паттерны: номера через запятую"
	questionProjectIdentifier = "Введите идентификатор проекта"
	questionConfirmSummary    = "Записать начальную структуру проекта? (да/нет)"
)

// optionNoPatterns — пункт списка паттернов, означающий пустой набор. Пустой набор выбирается
// пунктом, а не пустым ответом: пустой ответ переспрашивает любой вопрос опроса.
const optionNoPatterns = "без паттернов"

// reasonPatternsConflict — причина отказа, когда пункт «без паттернов» выбран вместе с другими.
const reasonPatternsConflict = "Пункт «без паттернов» не выбирается вместе с другими пунктами."

// CollectProfileOutput — выходной контракт сценария сбора профиля: собранный профиль,
// идентификатор проекта и ответ в сводке. Отказ в сводке — решение человека, поэтому приходит
// значением, а не ошибкой.
type CollectProfileOutput struct {
	Profile           domain.Profile
	ProjectIdentifier string
	Confirmed         bool
}

// CollectProfileUseCase — сценарий сбора профиля: порядок секций опроса, создание профиля
// доменным конструктором и сводка как последняя точка отказа. Входного контракта у сценария нет:
// профиль задаётся только опросом, и ни аргументы, ни окружение в него не участвуют.
type CollectProfileUseCase interface {
	Execute() (CollectProfileOutput, error)
}

// collectProfile — реализация сценария сбора профиля.
type collectProfile struct {
	prompter Prompter
}

// NewCollectProfile возвращает сценарий сбора профиля.
func NewCollectProfile(prompter Prompter) CollectProfileUseCase {
	return collectProfile{prompter: prompter}
}

// Execute проводит опрос в порядке «язык, тип проекта, архитектура, паттерны, идентификатор
// проекта» и предъявляет сводку. Ранее данные ответы заново не спрашиваются: переспрашивается
// только тот вопрос, ответ на который не годится.
func (c collectProfile) Execute() (CollectProfileOutput, error) {
	language, err := selectValue(c.prompter, questionLanguage, domain.Languages())
	if err != nil {
		return CollectProfileOutput{}, err
	}

	projectType, err := selectValue(c.prompter, questionProjectType, domain.ProjectTypes())
	if err != nil {
		return CollectProfileOutput{}, err
	}

	architecture, err := selectValue(c.prompter, questionArchitecture, domain.Architectures())
	if err != nil {
		return CollectProfileOutput{}, err
	}

	patterns, err := c.selectPatterns()
	if err != nil {
		return CollectProfileOutput{}, err
	}

	profile, err := domain.NewProfile(language, projectType, architecture, patterns)
	if err != nil {
		return CollectProfileOutput{}, fmt.Errorf("создание профиля по ответам опроса: %w", err)
	}

	identifier, err := c.readProjectIdentifier()
	if err != nil {
		return CollectProfileOutput{}, err
	}

	answer, err := c.prompter.ReadLine(summaryPrompt(profile, identifier))
	if err != nil {
		return CollectProfileOutput{}, err
	}

	return CollectProfileOutput{
		Profile:           profile,
		ProjectIdentifier: identifier,
		Confirmed:         confirms(answer),
	}, nil
}

// selectPatterns задаёт вопрос о паттернах: несколько номеров через запятую либо пункт «без
// паттернов». Пункт пустого набора вместе с другими номерами — неразобранный ответ: причина
// печатается, и тот же вопрос задаётся снова.
func (c collectProfile) selectPatterns() ([]domain.Pattern, error) {
	dictionary := domain.Patterns()
	items := append(options(dictionary), optionNoPatterns)
	noPatterns := len(dictionary)
	prompt := questionPatterns

	for {
		indexes, err := c.prompter.SelectMany(prompt, items)
		if err != nil {
			return nil, err
		}

		if slices.Contains(indexes, noPatterns) {
			if len(indexes) == 1 {
				return nil, nil
			}

			prompt = reasonPatternsConflict + "\n" + questionPatterns

			continue
		}

		patterns := make([]domain.Pattern, 0, len(indexes))
		for _, index := range indexes {
			patterns = append(patterns, dictionary[index])
		}

		return patterns, nil
	}
}

// readProjectIdentifier спрашивает идентификатор проекта. Умолчания у вопроса нет: ответ,
// нарушающий правило домена, печатается причиной и повторяет тот же вопрос, а число попыток
// не ограничивается.
func (c collectProfile) readProjectIdentifier() (string, error) {
	prompt := questionProjectIdentifier

	for {
		answer, err := c.prompter.ReadLine(prompt)
		if err != nil {
			return "", err
		}

		if invalid := domain.ValidateProjectIdentifier(answer); invalid != nil {
			prompt = invalid.Error() + "\n" + questionProjectIdentifier

			continue
		}

		return answer, nil
	}
}

// selectValue задаёт вопрос об измерении с одиночным выбором и возвращает выбранное значение
// словаря. Разбор ответа и повтор вопроса при неразобранном ответе — дело контракта опроса.
func selectValue[T ~string](prompter Prompter, question string, dictionary []T) (T, error) {
	var empty T

	index, err := prompter.SelectOne(question, options(dictionary))
	if err != nil {
		return empty, err
	}

	if index < 0 || index >= len(dictionary) {
		return empty, fmt.Errorf("номер пункта %d вне словаря измерения", index)
	}

	return dictionary[index], nil
}

// summaryPrompt собирает сводку: собранный профиль, идентификатор проекта и вопрос
// о подтверждении. План генерации в сводку приходит вместе с конвейером генерации.
func summaryPrompt(profile domain.Profile, identifier string) string {
	patterns := optionNoPatterns
	if selected := profile.Patterns(); len(selected) > 0 {
		patterns = strings.Join(options(selected), ", ")
	}

	lines := []string{
		"Сводка:",
		"  язык:                  " + string(profile.Language()),
		"  тип проекта:           " + string(profile.ProjectType()),
		"  архитектура:           " + string(profile.Architecture()),
		"  паттерны:              " + patterns,
		"  идентификатор проекта: " + identifier,
		"",
		questionConfirmSummary,
	}

	return strings.Join(lines, "\n")
}
