package app_test

import (
	"errors"
	"testing"

	planapp "github.com/united-software-platform/extruder/internal/plan/app"
	plan "github.com/united-software-platform/extruder/internal/plan/domain"
	profile "github.com/united-software-platform/extruder/internal/profile/domain"
)

// dimensions собирает три измерения, которые принимает построение плана.
func dimensions(t *testing.T, projectType, architecture string, patterns ...string) (
	profile.ProjectType, profile.Architecture, profile.Patterns,
) {
	t.Helper()
	pt, err := profile.NewProjectType(projectType)
	if err != nil {
		t.Fatalf("тип проекта: %v", err)
	}
	arch, err := profile.NewArchitecture(architecture)
	if err != nil {
		t.Fatalf("архитектура: %v", err)
	}
	set, err := profile.NewPatterns(patterns)
	if err != nil {
		t.Fatalf("паттерны: %v", err)
	}
	return pt, arch, set
}

func buildFirstSlice(t *testing.T) plan.Plan {
	t.Helper()
	pt, arch, set := dimensions(t, "service", "layered")
	structure, err := planapp.BuildPlan(pt, arch, set)
	if err != nil {
		t.Fatalf("построение плана: %v", err)
	}
	return structure
}

func TestServiceProjectTypeGivesOneModule(t *testing.T) {
	structure := buildFirstSlice(t)
	if len(structure.Modules) != 1 {
		t.Fatalf("модулей %d, ожидался 1", len(structure.Modules))
	}
}

func TestArchitectureDefinesLayers(t *testing.T) {
	structure := buildFirstSlice(t)
	kinds := make([]plan.LayerKind, 0, len(structure.Modules[0].Layers))
	for _, layer := range structure.Modules[0].Layers {
		kinds = append(kinds, layer.Kind)
	}
	expected := []plan.LayerKind{plan.LayerDomain, plan.LayerApp, plan.LayerInfra}
	if len(kinds) != len(expected) {
		t.Fatalf("слоёв %d, ожидалось %d: %v", len(kinds), len(expected), kinds)
	}
	for i := range expected {
		if kinds[i] != expected[i] {
			t.Errorf("слой %d — %q, ожидался %q", i, kinds[i], expected[i])
		}
	}
}

func TestProjectRolesAreOutsideLayers(t *testing.T) {
	structure := buildFirstSlice(t)
	expected := []plan.Role{plan.RoleModuleManifest, plan.RoleEntrypoint, plan.RoleBuildTarget}
	if len(structure.ProjectRoles) != len(expected) {
		t.Fatalf("ролей проекта %d, ожидалось %d: %v",
			len(structure.ProjectRoles), len(expected), structure.ProjectRoles)
	}
	for i := range expected {
		if structure.ProjectRoles[i] != expected[i] {
			t.Errorf("роль проекта %d — %q, ожидалась %q", i, structure.ProjectRoles[i], expected[i])
		}
	}
	for _, layer := range structure.Modules[0].Layers {
		for _, role := range layer.Roles {
			for _, projectRole := range expected {
				if role == projectRole {
					t.Errorf("роль проекта %q попала в слой %q", role, layer.Kind)
				}
			}
		}
	}
}

func TestFirstSliceMatchesExpectedStructure(t *testing.T) {
	structure := buildFirstSlice(t)
	expected := map[plan.LayerKind][]plan.Role{
		plan.LayerDomain: {plan.RoleEntity, plan.RoleRepositoryContract},
		plan.LayerApp:    {plan.RoleUseCase},
		plan.LayerInfra:  {plan.RoleRepositoryImpl, plan.RoleInboundHandler},
	}
	for _, layer := range structure.Modules[0].Layers {
		want := expected[layer.Kind]
		if len(layer.Roles) != len(want) {
			t.Errorf("слой %q: ролей %d, ожидалось %d — %v", layer.Kind, len(layer.Roles), len(want), layer.Roles)
			continue
		}
		for i := range want {
			if layer.Roles[i] != want[i] {
				t.Errorf("слой %q, роль %d — %q, ожидалась %q", layer.Kind, i, layer.Roles[i], want[i])
			}
		}
	}
}

func TestEmptyPatternsAddNoRoles(t *testing.T) {
	withoutPatterns := buildFirstSlice(t)
	pt, arch, set := dimensions(t, "service", "layered", "ddd", "cqrs")
	withPatterns, err := planapp.BuildPlan(pt, arch, set)
	if err != nil {
		t.Fatalf("построение плана с паттернами: %v", err)
	}
	for i, layer := range withoutPatterns.Modules[0].Layers {
		if len(layer.Roles) > len(withPatterns.Modules[0].Layers[i].Roles) {
			t.Errorf("слой %q: пустой набор паттернов дал больше ролей, чем непустой", layer.Kind)
		}
	}
	if len(withoutPatterns.ProjectRoles) != len(withPatterns.ProjectRoles) {
		t.Error("набор паттернов изменил роли уровня проекта")
	}
}

func TestBuildPlanIsDeterministic(t *testing.T) {
	first := buildFirstSlice(t)
	second := buildFirstSlice(t)
	if len(first.Modules) != len(second.Modules) {
		t.Fatal("число модулей различается между вызовами")
	}
	for m := range first.Modules {
		for l, layer := range first.Modules[m].Layers {
			other := second.Modules[m].Layers[l]
			if layer.Kind != other.Kind {
				t.Errorf("порядок слоёв различается: %q против %q", layer.Kind, other.Kind)
			}
			for r := range layer.Roles {
				if layer.Roles[r] != other.Roles[r] {
					t.Errorf("порядок ролей различается в слое %q", layer.Kind)
				}
			}
		}
	}
}

func TestAllRolesOfPlanAreKnown(t *testing.T) {
	structure := buildFirstSlice(t)
	for _, layer := range structure.Modules[0].Layers {
		if !layer.Kind.IsKnown() {
			t.Errorf("вид слоя %d не опознан", int(layer.Kind))
		}
		for _, role := range layer.Roles {
			if !role.IsKnown() {
				t.Errorf("роль %d в слое %q не опознана", int(role), layer.Kind)
			}
		}
	}
	for _, role := range structure.ProjectRoles {
		if !role.IsKnown() {
			t.Errorf("роль проекта %d не опознана", int(role))
		}
	}
}

// Перечни допустимых значений живут в профиле, а их отображение в план — здесь. Тест не даёт
// им разойтись: значение, допущенное профилем, но не поддержанное планом, — дефект.
func TestEveryValueAllowedByProfileIsSupportedByPlan(t *testing.T) {
	for _, projectType := range profile.AllowedValues(profile.DimensionProjectType) {
		for _, architecture := range profile.AllowedValues(profile.DimensionArchitecture) {
			pt, arch, set := dimensions(t, projectType, architecture)
			if _, err := planapp.BuildPlan(pt, arch, set); err != nil {
				t.Errorf("профиль %q + %q планом не поддержан: %v", projectType, architecture, err)
			}
		}
	}
	pt, arch, _ := dimensions(t, "service", "layered")
	for _, pattern := range profile.AllowedValues(profile.DimensionPatterns) {
		set, err := profile.NewPatterns([]string{pattern})
		if err != nil {
			t.Fatalf("паттерн %q: %v", pattern, err)
		}
		if _, err := planapp.BuildPlan(pt, arch, set); err != nil {
			t.Errorf("паттерн %q планом не поддержан: %v", pattern, err)
		}
	}
}

func TestBuildPlanRejectsValueUnknownToPlan(t *testing.T) {
	var unsupported planapp.UnsupportedValueError
	_, err := planapp.BuildPlan(profile.ProjectType{}, profile.Architecture{}, profile.Patterns{})
	if !errors.As(err, &unsupported) {
		t.Fatalf("ожидалась UnsupportedValueError, получено: %v", err)
	}
}
