// Package cli — точка сборки инструмента: разбирает аргументы, вызывает сценарии модулей
// и предъявляет результат.
//
// Пакет лежит вне модулей намеренно. Он знает и «профиль», и «план структуры», а модули друг
// о друге не знают: иначе «профиль» зависел бы от «плана» ради одной печати, и межмодульная
// зависимость пошла бы в обход контрактов домена [CLAR-023].
package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	planapp "github.com/united-software-platform/extruder/internal/plan/app"
	planinfra "github.com/united-software-platform/extruder/internal/plan/infra"
	profileapp "github.com/united-software-platform/extruder/internal/profile/app"
	profiledomain "github.com/united-software-platform/extruder/internal/profile/domain"
)

// Имена флагов: по одному на измерение. Независимость измерений видна прямо в синтаксисе —
// каждое измерение задаётся своим флагом, порядок флагов безразличен.
const (
	flagLanguage     = "language"
	flagProjectType  = "project-type"
	flagArchitecture = "architecture"
	flagPatterns     = "patterns"
)

// NewCommand собирает команду разбора профиля. Зависимость от библиотеки разбора
// командной строки заканчивается здесь: сценарий приложения о ней не знает.
func NewCommand() *cobra.Command {
	var (
		language     string
		projectType  string
		architecture string
		patterns     []string
	)

	command := &cobra.Command{
		Use:   "extruder",
		Short: "Превращает архитектурный профиль приложения в его начальную структуру",
		Long:  longDescription(),
		// Диагностику печатает вызывающая сторона: иначе cobra отправит её в поток вывода,
		// где по требованию спецификации при отказе не должно быть ничего
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(command *cobra.Command, _ []string) error {
			output, err := profileapp.NewParseProfileUseCase().Execute(profileapp.ParseProfileInput{
				Language:     language,
				ProjectType:  projectType,
				Architecture: architecture,
				Patterns:     patterns,
			})
			if err != nil {
				return err
			}
			out := command.OutOrStdout()
			if err := writeProfile(out, output.Profile); err != nil {
				return err
			}
			return writePlan(out, output.Profile)
		},
	}

	flags := command.Flags()
	flags.StringVar(&language, flagLanguage, "", flagUsage(profiledomain.DimensionLanguage, true))
	flags.StringVar(&projectType, flagProjectType, "", flagUsage(profiledomain.DimensionProjectType, true))
	flags.StringVar(&architecture, flagArchitecture, "", flagUsage(profiledomain.DimensionArchitecture, true))
	flags.StringSliceVar(&patterns, flagPatterns, nil, flagUsage(profiledomain.DimensionPatterns, false))

	return command
}

// flagUsage собирает подсказку флага из перечня домена: перечень значений в справке
// и перечень, по которому идёт проверка, — один и тот же источник.
func flagUsage(dimension profiledomain.Dimension, required bool) string {
	obligation := "необязательно"
	if required {
		obligation = "обязательно"
	}
	return fmt.Sprintf(
		"измерение %q (%s); допустимые значения: %s",
		string(dimension), obligation, strings.Join(profiledomain.AllowedValues(dimension), ", "),
	)
}

// longDescription — текст справки. Перечень измерений с их обязательностью и допустимыми
// значениями берётся из домена, а не дублируется здесь.
func longDescription() string {
	var builder strings.Builder
	builder.WriteString("Extruder превращает архитектурный профиль приложения в его начальную структуру.\n\n")
	builder.WriteString("Профиль собирается из четырёх независимых измерений:\n")
	rows := []struct {
		dimension profiledomain.Dimension
		required  bool
	}{
		{profiledomain.DimensionLanguage, true},
		{profiledomain.DimensionProjectType, true},
		{profiledomain.DimensionArchitecture, true},
		{profiledomain.DimensionPatterns, false},
	}
	for _, row := range rows {
		obligation := "необязательно"
		if row.required {
			obligation = "обязательно"
		}
		builder.WriteString(fmt.Sprintf(
			"  %-20s %-14s допустимые значения: %s\n",
			string(row.dimension), obligation,
			strings.Join(profiledomain.AllowedValues(row.dimension), ", "),
		))
	}
	return builder.String()
}

// writeProfile предъявляет разобранный профиль: все четыре измерения с их значениями,
// включая пустой набор паттернов.
func writeProfile(out io.Writer, profile profiledomain.Profile) error {
	patterns := "—"
	if !profile.Patterns().IsEmpty() {
		patterns = strings.Join(profile.Patterns().Values(), ", ")
	}
	rows := [][2]string{
		{string(profiledomain.DimensionLanguage), profile.Language().String()},
		{string(profiledomain.DimensionProjectType), profile.ProjectType().String()},
		{string(profiledomain.DimensionArchitecture), profile.Architecture().String()},
		{string(profiledomain.DimensionPatterns), patterns},
	}
	for _, row := range rows {
		if _, err := fmt.Fprintf(out, "%-20s %s\n", row[0]+":", row[1]); err != nil {
			return err
		}
	}
	return nil
}

// writePlan строит план структуры и предъявляет его после профиля.
//
// Построение получает три измерения, а не профиль целиком: измерение «язык» в фазу построения
// не передаётся, и сослаться на него там не на что.
func writePlan(out io.Writer, profile profiledomain.Profile) error {
	structure, err := planapp.BuildPlan(
		profile.ProjectType(),
		profile.Architecture(),
		profile.Patterns(),
	)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out); err != nil {
		return err
	}
	return planinfra.Write(out, structure)
}

// Run выполняет команду и возвращает код возврата процесса. Логика запуска живёт здесь,
// а не в точке входа: иначе поведение при вызове без аргументов и неизменность файловой
// системы нечем проверить тестом.
func Run(args []string, stdout, stderr io.Writer) int {
	command := NewCommand()
	command.SetOut(stdout)
	command.SetErr(stderr)
	command.SetArgs(args)

	// Вызов без единого аргумента: показать справку и завершиться отказом. Справка идёт
	// в поток ошибок, потому что поток вывода при отказе обязан остаться пустым —
	// в нём печатается только разобранный профиль.
	if len(args) == 0 {
		command.SetOut(stderr)
		if err := command.Help(); err != nil {
			// Отказ записи в поток ошибок сообщать уже некуда: остаётся код возврата
			_, _ = fmt.Fprintln(stderr, err)
		}
		return 1
	}

	if err := command.Execute(); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
