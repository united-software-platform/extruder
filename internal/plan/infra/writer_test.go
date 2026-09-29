package infra_test

import (
	"bytes"
	"strings"
	"testing"

	planapp "github.com/united-software-platform/extruder/internal/plan/app"
	plan "github.com/united-software-platform/extruder/internal/plan/domain"
	planinfra "github.com/united-software-platform/extruder/internal/plan/infra"
	profile "github.com/united-software-platform/extruder/internal/profile/domain"
)

// эталон плана первого вертикального среза: Go, сервис, Layered, без паттернов.
const firstSliceReference = `План структуры:
модуль
├── слой domain
│   ├── роль entity
│   └── роль repository-contract
├── слой app
│   └── роль use-case
└── слой infra
    ├── роль repository-impl
    └── роль inbound-handler
роль module-manifest
роль entrypoint
роль build-target
`

func renderFirstSlice(t *testing.T, patterns ...string) string {
	t.Helper()
	projectType, err := profile.NewProjectType("service")
	if err != nil {
		t.Fatalf("тип проекта: %v", err)
	}
	architecture, err := profile.NewArchitecture("layered")
	if err != nil {
		t.Fatalf("архитектура: %v", err)
	}
	set, err := profile.NewPatterns(patterns)
	if err != nil {
		t.Fatalf("паттерны: %v", err)
	}
	structure, err := planapp.BuildPlan(projectType, architecture, set)
	if err != nil {
		t.Fatalf("построение плана: %v", err)
	}
	var out bytes.Buffer
	if err := planinfra.Write(&out, structure); err != nil {
		t.Fatalf("печать плана: %v", err)
	}
	return out.String()
}

func TestWriteMatchesFirstSliceReference(t *testing.T) {
	got := renderFirstSlice(t)
	if got != firstSliceReference {
		t.Errorf("план разошёлся с эталоном.\nполучено:\n%s\nожидалось:\n%s", got, firstSliceReference)
	}
}

func TestWriteShowsHierarchy(t *testing.T) {
	got := renderFirstSlice(t)
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")

	var moduleSeen, layerSeen bool
	for _, line := range lines {
		switch {
		case line == "модуль":
			moduleSeen = true
		case strings.Contains(line, "слой "):
			if !moduleSeen {
				t.Error("слой предъявлен раньше модуля")
			}
			layerSeen = true
		case strings.Contains(line, "роль ") && strings.HasPrefix(line, " ") ||
			strings.Contains(line, "роль ") && strings.ContainsAny(line, "├└│"):
			if !layerSeen {
				t.Error("роль слоя предъявлена раньше слоя")
			}
		}
	}
	if !moduleSeen || !layerSeen {
		t.Error("в выводе нет модуля или слоя")
	}
}

func TestWritePutsProjectRolesOutsideLayers(t *testing.T) {
	got := renderFirstSlice(t)
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")

	for _, role := range []plan.Role{plan.RoleModuleManifest, plan.RoleEntrypoint, plan.RoleBuildTarget} {
		expected := "роль " + role.String()
		found := false
		for _, line := range lines {
			if line == expected {
				found = true
			}
		}
		if !found {
			t.Errorf("роль уровня проекта %q не предъявлена на верхнем уровне", role)
		}
	}
}

// Узлами считаются строки плана: модуль, слой и роль. Заголовок к ним не относится.
func nodeLines(t *testing.T, rendered string) []string {
	t.Helper()
	var nodes []string
	for _, line := range strings.Split(strings.TrimRight(rendered, "\n"), "\n") {
		if line == "План структуры:" {
			continue
		}
		nodes = append(nodes, line)
	}
	if len(nodes) == 0 {
		t.Fatal("в выводе нет ни одного узла")
	}
	return nodes
}

func TestNodesCarryNoLanguageMarkers(t *testing.T) {
	for _, patterns := range [][]string{nil, {"ddd"}, {"ddd", "cqrs"}} {
		rendered := renderFirstSlice(t, patterns...)
		for _, line := range nodeLines(t, rendered) {
			if strings.Contains(line, ".") {
				t.Errorf("узел содержит точку — признак расширения: %q", line)
			}
			if strings.Contains(line, "/") || strings.Contains(line, "\\") {
				t.Errorf("узел содержит разделитель пути: %q", line)
			}
			if strings.Contains(line, "package") {
				t.Errorf("узел содержит объявление пакета: %q", line)
			}
		}
	}
}

func TestNodesCarryNoContent(t *testing.T) {
	for _, line := range nodeLines(t, renderFirstSlice(t)) {
		if strings.Contains(line, "{") || strings.Contains(line, "(") || strings.Contains(line, "=") {
			t.Errorf("узел похож на содержимое файла: %q", line)
		}
	}
}

func TestWriteIsDeterministic(t *testing.T) {
	first := renderFirstSlice(t)
	second := renderFirstSlice(t)
	if first != second {
		t.Errorf("повторная печать дала другой результат:\n%s\nпротив\n%s", first, second)
	}
}
