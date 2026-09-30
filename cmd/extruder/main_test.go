package main

import (
	"bytes"
	"regexp"
	"testing"
)

// versionPattern — формат строки версии: X.Y.Z с необязательным суффиксом предрелиза.
var versionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$`)

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

func TestRunWritesVersion(t *testing.T) {
	var out bytes.Buffer

	if err := run(&out); err != nil {
		t.Fatalf("run вернула ошибку: %v", err)
	}

	if got := out.String(); got != Version()+"\n" {
		t.Fatalf("напечатано %q, ожидалась строка версии %q", got, Version())
	}
}
