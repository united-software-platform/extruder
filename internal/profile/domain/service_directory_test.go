package domain

import "testing"

func TestBlocksEmptiness(t *testing.T) {
	cases := []struct {
		name  string
		entry string
		want  bool
	}{
		{name: "служебный каталог из перечня", entry: ".git", want: false},
		{name: "служебный каталог редактора", entry: ".idea", want: false},
		{name: "скрытый файл вне перечня", entry: ".gitignore", want: true},
		{name: "обычный файл", entry: "README.md", want: true},
		{name: "обычный каталог", entry: "cmd", want: true},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := BlocksEmptiness(test.entry); got != test.want {
				t.Errorf("BlocksEmptiness(%q) = %t, ожидалось %t", test.entry, got, test.want)
			}
		})
	}
}

func TestServiceDirectoriesIsClosedList(t *testing.T) {
	want := []string{".git", ".hg", ".svn", ".idea", ".vscode", ".DS_Store"}

	got := ServiceDirectories()
	if len(got) != len(want) {
		t.Fatalf("перечень служебных каталогов %v, ожидался %v", got, want)
	}

	for index, name := range want {
		if got[index] != name {
			t.Errorf("служебный каталог %d — %q, ожидался %q", index, got[index], name)
		}
	}
}
