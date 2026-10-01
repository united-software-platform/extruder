package application

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestCheckTargetDirRejectsUnsuitableDir(t *testing.T) {
	target := &targetDirDouble{
		writableErr: fmt.Errorf("%w: каталога %q не существует", ErrTargetDirUnsuitable, "/project"),
	}
	prompter := &promptDouble{}

	_, err := NewCheckTargetDir(target, prompter).Execute(CheckTargetDirInput{Dir: "/project"})
	if err == nil {
		t.Fatal("непригодный каталог принят, ожидался отказ")
	}

	if !errors.Is(err, ErrTargetDirUnsuitable) {
		t.Errorf("ошибка %v не сравнима с ErrTargetDirUnsuitable", err)
	}

	if len(prompter.prompts) != 0 {
		t.Errorf("задано вопросов %d, ожидалось ни одного: %v", len(prompter.prompts), prompter.prompts)
	}

	if len(target.listed) != 0 {
		t.Errorf("содержимое непригодного каталога перечислялось: %v", target.listed)
	}
}

func TestCheckTargetDirPassesOnEmptyDir(t *testing.T) {
	target := &targetDirDouble{}
	prompter := &promptDouble{}

	out, err := NewCheckTargetDir(target, prompter).Execute(CheckTargetDirInput{Dir: "/project"})
	if err != nil {
		t.Fatalf("проверка пустого каталога вернула ошибку: %v", err)
	}

	if !out.Proceed {
		t.Error("пустой каталог не пропущен")
	}

	if len(prompter.prompts) != 0 {
		t.Errorf("на пустом каталоге задано вопросов %d, ожидалось ни одного", len(prompter.prompts))
	}
}

func TestCheckTargetDirPassesOnServiceDirectoriesOnly(t *testing.T) {
	target := &targetDirDouble{entries: []string{".git", ".idea"}}
	prompter := &promptDouble{}

	out, err := NewCheckTargetDir(target, prompter).Execute(CheckTargetDirInput{Dir: "/project"})
	if err != nil {
		t.Fatalf("проверка каталога со служебными каталогами вернула ошибку: %v", err)
	}

	if !out.Proceed {
		t.Error("каталог со служебными каталогами не пропущен")
	}

	if len(prompter.prompts) != 0 {
		t.Errorf("задан вопрос о содержимом: %v", prompter.prompts)
	}
}

func TestCheckTargetDirAsksOnceOnNonEmptyDir(t *testing.T) {
	target := &targetDirDouble{entries: []string{".git", "README.md", "cmd", ".gitignore"}}
	prompter := &promptDouble{lines: []string{"да"}}

	out, err := NewCheckTargetDir(target, prompter).Execute(CheckTargetDirInput{Dir: "/project"})
	if err != nil {
		t.Fatalf("проверка непустого каталога вернула ошибку: %v", err)
	}

	if !out.Proceed {
		t.Error("ответ «да» не пропустил запуск дальше")
	}

	if len(prompter.prompts) != 1 {
		t.Fatalf("задано вопросов %d, ожидался ровно один: %v", len(prompter.prompts), prompter.prompts)
	}

	prompt := prompter.prompts[0]
	for _, path := range []string{"README.md", "cmd", ".gitignore"} {
		if !strings.Contains(prompt, path) {
			t.Errorf("в перечне нет пути %q: %s", path, prompt)
		}
	}

	if strings.Contains(prompt, ".git\n") || strings.Contains(prompt, "/.git\n") {
		t.Errorf("служебный каталог .git попал в перечень: %s", prompt)
	}
}

func TestCheckTargetDirDeclinesOnEmptyAnswer(t *testing.T) {
	target := &targetDirDouble{entries: []string{"README.md"}}
	prompter := &promptDouble{lines: []string{""}}

	out, err := NewCheckTargetDir(target, prompter).Execute(CheckTargetDirInput{Dir: "/project"})
	if err != nil {
		t.Fatalf("проверка непустого каталога вернула ошибку: %v", err)
	}

	if out.Proceed {
		t.Error("пустой ответ пропустил запуск дальше, ожидался отказ")
	}
}

func TestCheckTargetDirDeclinesOnNo(t *testing.T) {
	target := &targetDirDouble{entries: []string{"README.md"}}
	prompter := &promptDouble{lines: []string{"нет"}}

	out, err := NewCheckTargetDir(target, prompter).Execute(CheckTargetDirInput{Dir: "/project"})
	if err != nil {
		t.Fatalf("проверка непустого каталога вернула ошибку: %v", err)
	}

	if out.Proceed {
		t.Error("ответ «нет» пропустил запуск дальше, ожидался отказ")
	}
}

func TestCheckTargetDirPropagatesInterruptedInput(t *testing.T) {
	target := &targetDirDouble{entries: []string{"README.md"}}
	prompter := &promptDouble{}

	_, err := NewCheckTargetDir(target, prompter).Execute(CheckTargetDirInput{Dir: "/project"})
	if !errors.Is(err, ErrInputInterrupted) {
		t.Errorf("ошибка %v не сравнима с ErrInputInterrupted", err)
	}
}
