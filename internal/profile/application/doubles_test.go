package application

// Двойники контрактов слоя приложения: контракты компилируются вместе с ними, а сценарии
// проверяются без потоков ввода-вывода и файловой системы.
var (
	_ Prompter        = (*promptDouble)(nil)
	_ TargetDirectory = (*targetDirDouble)(nil)
)

// promptDouble — двойник контракта опроса: отдаёт заранее заданные ответы по порядку
// и запоминает подсказки, с которыми были заданы вопросы. Исчерпанный сценарий ответов
// означает конец потока ввода.
type promptDouble struct {
	single   []int
	multiple [][]int
	lines    []string
	prompts  []string
}

// SelectOne отдаёт очередной ответ на вопрос с одиночным выбором.
func (d *promptDouble) SelectOne(prompt string, _ []string) (int, error) {
	d.prompts = append(d.prompts, prompt)

	if len(d.single) == 0 {
		return 0, ErrInputInterrupted
	}

	answer := d.single[0]
	d.single = d.single[1:]

	return answer, nil
}

// SelectMany отдаёт очередной ответ на вопрос с множественным выбором.
func (d *promptDouble) SelectMany(prompt string, _ []string) ([]int, error) {
	d.prompts = append(d.prompts, prompt)

	if len(d.multiple) == 0 {
		return nil, ErrInputInterrupted
	}

	answer := d.multiple[0]
	d.multiple = d.multiple[1:]

	return answer, nil
}

// ReadLine отдаёт очередную строку ответа.
func (d *promptDouble) ReadLine(prompt string) (string, error) {
	d.prompts = append(d.prompts, prompt)

	if len(d.lines) == 0 {
		return "", ErrInputInterrupted
	}

	answer := d.lines[0]
	d.lines = d.lines[1:]

	return answer, nil
}

// targetDirDouble — двойник контракта целевого каталога: отдаёт заданное содержимое и заданный
// исход проверки пригодности, запоминая каталоги, по которым его спрашивали.
type targetDirDouble struct {
	entries     []string
	writableErr error
	entriesErr  error
	ensured     []string
	listed      []string
}

// EnsureWritable отдаёт заданный исход проверки пригодности.
func (d *targetDirDouble) EnsureWritable(dir string) error {
	d.ensured = append(d.ensured, dir)

	return d.writableErr
}

// Entries отдаёт заданное содержимое каталога.
func (d *targetDirDouble) Entries(dir string) ([]string, error) {
	d.listed = append(d.listed, dir)

	if d.entriesErr != nil {
		return nil, d.entriesErr
	}

	return d.entries, nil
}
