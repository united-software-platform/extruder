package main

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// versionPattern — формат строки версии: X.Y.Z с необязательным суффиксом предрелиза.
var versionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$`)

// Ответы полного прохода опроса: первые пункты трёх измерений, пункт «без паттернов»,
// идентификатор проекта и подтверждение в сводке.
const (
	fullPassAnswers    = "1\n1\n1\n3\norders\nда\n"
	declinedAnswers    = "1\n1\n1\n3\norders\nнет\n"
	interruptedAnswers = "1\n1\n"
)

// Опознавательные части вывода. Проверяется не вся строка, а то, по чему исход узнаётся:
// формулировка сообщения может меняться, класс исхода — нет.
const (
	markerLanguageQuestion = "Выберите язык"
	markerNonEmptyDir      = "не пуст"
	markerInteractiveFlag  = "-it"
)

// execute запускает точку входа на подставленном потоке ответов и каталоге-параметре.
func execute(t *testing.T, answers, dir string, terminal bool) (int, string, *strings.Reader) {
	t.Helper()

	in := strings.NewReader(answers)

	var out bytes.Buffer

	code := run(in, &out, dir, terminal)

	return code, out.String(), in
}

// firstLine возвращает первую строку вывода.
func firstLine(printed string) string {
	line, _, _ := strings.Cut(printed, "\n")

	return line
}

func TestVersionIsNotEmpty(t *testing.T) {
	if Version() == "" {
		t.Fatal("строка версии пуста")
	}
}

func TestVersionMatchesFormat(t *testing.T) {
	got := Version()
	if !versionPattern.MatchString(got) {
		t.Fatalf("строка версии %q не соответствует формату X.Y.Z с необязательным предрелизом", got)
	}
}

func TestRunPrintsVersionAsFirstLine(t *testing.T) {
	cases := []struct {
		name     string
		answers  string
		dir      func(t *testing.T) string
		terminal bool
	}{
		{
			name:     "полный проход опроса",
			answers:  fullPassAnswers,
			dir:      func(t *testing.T) string { t.Helper(); return t.TempDir() },
			terminal: true,
		},
		{
			name:     "отказ из-за отсутствия терминала",
			answers:  "",
			dir:      func(t *testing.T) string { t.Helper(); return t.TempDir() },
			terminal: false,
		},
		{
			name:    "отказ из-за непригодного целевого каталога",
			answers: "",
			dir: func(t *testing.T) string {
				t.Helper()

				return filepath.Join(t.TempDir(), "не-смонтирован")
			},
			terminal: true,
		},
		{
			name:     "отказ из-за прерванного ввода",
			answers:  interruptedAnswers,
			dir:      func(t *testing.T) string { t.Helper(); return t.TempDir() },
			terminal: true,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, printed, _ := execute(t, test.answers, test.dir(t), test.terminal)

			if got := firstLine(printed); got != Version() {
				t.Errorf("первая строка вывода %q, ожидалась строка версии %q", got, Version())
			}
		})
	}
}

func TestRunCompletesFullPass(t *testing.T) {
	code, printed, _ := execute(t, fullPassAnswers, t.TempDir(), true)

	if code != exitOK {
		t.Fatalf("код возврата %d, ожидался %d:\n%s", code, exitOK, printed)
	}

	for _, marker := range []string{markerLanguageQuestion, "Сводка:", "go", "monolith", "clean-architecture", "orders"} {
		if !strings.Contains(printed, marker) {
			t.Errorf("в выводе нет %q:\n%s", marker, printed)
		}
	}
}

func TestRunRejectsWithoutTerminal(t *testing.T) {
	code, printed, in := execute(t, fullPassAnswers, t.TempDir(), false)

	if code != exitNoTerminal {
		t.Fatalf("код возврата %d, ожидался %d:\n%s", code, exitNoTerminal, printed)
	}

	if !strings.Contains(printed, markerInteractiveFlag) {
		t.Errorf("отказ не называет команду запуска с %q:\n%s", markerInteractiveFlag, printed)
	}

	if strings.Contains(printed, markerLanguageQuestion) {
		t.Errorf("без терминала задан вопрос опроса:\n%s", printed)
	}

	if in.Len() != len(fullPassAnswers) {
		t.Error("без терминала инструмент читал стандартный ввод, ожидался отказ без ожидания ввода")
	}
}

func TestRunRejectsUnsuitableTargetDir(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "не-смонтирован")

	code, printed, in := execute(t, fullPassAnswers, missing, true)

	if code != exitTargetDirUnsuitable {
		t.Fatalf("код возврата %d, ожидался %d:\n%s", code, exitTargetDirUnsuitable, printed)
	}

	if strings.Contains(printed, markerLanguageQuestion) {
		t.Errorf("на непригодном каталоге задан вопрос опроса:\n%s", printed)
	}

	if in.Len() != len(fullPassAnswers) {
		t.Error("на непригодном каталоге инструмент читал стандартный ввод")
	}
}

func TestRunReportsInterruptedInput(t *testing.T) {
	code, printed, _ := execute(t, interruptedAnswers, t.TempDir(), true)

	if code != exitInterrupted {
		t.Fatalf("код возврата %d, ожидался %d:\n%s", code, exitInterrupted, printed)
	}

	if !strings.Contains(printed, "ввод прерван") {
		t.Errorf("отказ не назван «ввод прерван»:\n%s", printed)
	}
}

func TestRunDeclinedInSummary(t *testing.T) {
	dir := t.TempDir()

	code, printed, _ := execute(t, declinedAnswers, dir, true)

	if code != exitOK {
		t.Fatalf("код возврата %d, ожидался %d:\n%s", code, exitOK, printed)
	}

	if !strings.Contains(printed, messageDeclined) {
		t.Errorf("отказ пользователя не сообщён:\n%s", printed)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("чтение целевого каталога вернуло ошибку: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("целевой каталог изменён: %d элементов, ожидался пустой", len(entries))
	}
}

func TestRunDeclinedOnNonEmptyTargetDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("содержимое"), 0o600); err != nil {
		t.Fatalf("подготовка файла вернула ошибку: %v", err)
	}

	code, printed, _ := execute(t, "нет\n", dir, true)

	if code != exitOK {
		t.Fatalf("код возврата %d, ожидался %d:\n%s", code, exitOK, printed)
	}

	if !strings.Contains(printed, markerNonEmptyDir) {
		t.Errorf("непустой каталог не предъявлен:\n%s", printed)
	}

	if strings.Contains(printed, markerLanguageQuestion) {
		t.Errorf("после отказа на гейте начался опрос:\n%s", printed)
	}

	names, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("чтение целевого каталога вернуло ошибку: %v", err)
	}

	got := make([]string, 0, len(names))
	for _, entry := range names {
		got = append(got, entry.Name())
	}

	if want := []string{"README.md"}; !slices.Equal(got, want) {
		t.Errorf("содержимое целевого каталога %v, ожидалось %v: запуск изменил каталог", got, want)
	}
}

func TestRunAsksGateBeforeSurvey(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "README.md"), nil, 0o600); err != nil {
		t.Fatalf("подготовка файла вернула ошибку: %v", err)
	}

	_, printed, _ := execute(t, "да\n"+fullPassAnswers, dir, true)

	gate := strings.Index(printed, markerNonEmptyDir)
	survey := strings.Index(printed, markerLanguageQuestion)

	if gate < 0 || survey < 0 {
		t.Fatalf("в выводе нет гейта или первого вопроса опроса:\n%s", printed)
	}

	if gate > survey {
		t.Errorf("вопрос гейта задан после начала опроса:\n%s", printed)
	}

	if version := strings.Index(printed, Version()); version > gate {
		t.Errorf("версия напечатана после гейта:\n%s", printed)
	}
}
