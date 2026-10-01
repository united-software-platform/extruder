// Package main — точка входа Extruder: проверка терминала, гейт целевого каталога и опрос.
// Конвейер генерации и запись начальной структуры проекта приходят последующими изменениями.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/united-software-platform/extruder/internal/profile/application"
	"github.com/united-software-platform/extruder/internal/profile/infrastructure"
)

// version — версия инструмента. Файл версии и релизный сценарий появятся на этапе поставки,
// до тех пор строка фиксирована и означает невыпущенную версию.
const version = "0.0.0-dev"

// targetDir — целевой каталог: фиксированная точка монтирования внутри контейнера. Путь
// не вводится и не передаётся — пользователь выбирает каталог монтированием, а не вводом пути.
const targetDir = "/project"

// imageName — имя образа в команде запуска, которую называет отказ при отсутствии терминала.
// Полное имя в реестре образ получит на этапе поставки.
const imageName = "extruder"

// Коды возврата называют класс исхода запуска. Коды `4` и далее оставлены незанятыми: их занимают
// отказы конвейера генерации и записи, приходящие последующими изменениями.
const (
	// exitOK — запуск завершён штатно, включая отказ пользователя.
	exitOK = 0
	// exitNoTerminal — стандартный ввод не является терминалом.
	exitNoTerminal = 1
	// exitTargetDirUnsuitable — целевой каталог непригоден.
	exitTargetDirUnsuitable = 2
	// exitInterrupted — запуск прерван до записи: поток ввода кончился либо диалог
	// с пользователем оказался невозможен.
	exitInterrupted = 3
)

// Сообщения исходов запуска. Отказ пользователя и отказ окружения печатаются одинаково подробно:
// в образе без оболочки сообщение инструмента — единственный источник диагностики.
const (
	messageDeclined = "Запись не выполнена: целевой каталог не изменён."
	messageNotYet   = "Запись начальной структуры проекта в этой версии инструмента ещё не выполняется."
)

// Version возвращает строку версии инструмента.
func Version() string {
	return version
}

// run выполняет запуск целиком и возвращает его код. Поток ввода, поток вывода, целевой каталог
// и признак терминала приходят параметрами: порядок «версия → проверка терминала → гейт каталога
// → опрос → сводка» записан здесь, и проверяется он сквозным тестом, а не запуском контейнера.
func run(in io.Reader, out io.Writer, dir string, terminal bool) int {
	// Версия — первая строка любого запуска, включая отказы: в образе без оболочки это
	// единственный способ узнать, какой инструмент отработал.
	say(out, Version())

	if !terminal {
		say(out, noTerminalMessage(dir))

		return exitNoTerminal
	}

	prompter := infrastructure.NewPrompt(in, out)

	gate, err := application.NewCheckTargetDir(infrastructure.NewTargetDir(), prompter).
		Execute(application.CheckTargetDirInput{Dir: dir})
	if err != nil {
		return reject(out, err)
	}

	if !gate.Proceed {
		say(out, messageDeclined)

		return exitOK
	}

	collected, err := application.NewCollectProfile(prompter).Execute()
	if err != nil {
		return reject(out, err)
	}

	if !collected.Confirmed {
		say(out, messageDeclined)

		return exitOK
	}

	say(out, "Профиль принят: %s, идентификатор проекта %s.", collected.Profile, collected.ProjectIdentifier)
	say(out, messageNotYet)

	return exitOK
}

// reject печатает причину отказа и отображает её в код возврата. Отказ, не названный ни одним
// классом, завершает запуск кодом прерванного запуска: до записи дело не дошло, целевой каталог
// не изменён, а причину называет напечатанное сообщение.
func reject(out io.Writer, err error) int {
	say(out, "Отказ: %v", err)

	switch {
	case errors.Is(err, application.ErrTargetDirUnsuitable):
		return exitTargetDirUnsuitable
	case errors.Is(err, application.ErrInputInterrupted):
		return exitInterrupted
	default:
		return exitInterrupted
	}
}

// noTerminalMessage собирает отказ при отсутствии терминала: называет причину и команду запуска
// в интерактивном режиме. Ожидания ввода в этом случае не происходит — опрос не начинается.
func noTerminalMessage(dir string) string {
	return strings.Join([]string{
		"Отказ: стандартный ввод не является терминалом, а профиль задаётся только опросом.",
		"Запустите инструмент в интерактивном режиме:",
		fmt.Sprintf("  docker run -it --rm -v \"$PWD:%s\" %s", dir, imageName),
	}, "\n")
}

// say печатает строку в поток вывода. Ошибка записи намеренно не влияет на код возврата: код
// называет исход запуска, а о сломанном стандартном выводе сообщить всё равно некуда — в образе
// на `scratch` нет ни оболочки, ни журнала.
func say(out io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(out, format+"\n", args...)
}

func main() {
	os.Exit(run(os.Stdin, os.Stdout, targetDir, infrastructure.IsTerminal(os.Stdin)))
}
