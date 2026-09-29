package domain_test

import (
	"testing"

	plan "github.com/united-software-platform/extruder/internal/plan/domain"
)

func TestRoleNamesAreDefinedForEveryRole(t *testing.T) {
	for _, role := range plan.Roles() {
		if !role.IsKnown() {
			t.Errorf("роль %d не опознана, хотя входит в перечень", int(role))
		}
		if role.String() == "" {
			t.Errorf("у роли %d нет имени", int(role))
		}
	}
}

func TestRoleOutsideListIsNotKnown(t *testing.T) {
	outside := plan.Role(len(plan.Roles()) + 100)
	if outside.IsKnown() {
		t.Error("роль вне перечня опознана")
	}
	if outside.String() != "" {
		t.Errorf("у роли вне перечня есть имя %q", outside.String())
	}
	if plan.RoleUnknown.IsKnown() {
		t.Error("нулевое значение роли опознано")
	}
}

func TestRolesReturnsDeclarationOrder(t *testing.T) {
	all := plan.Roles()
	if len(all) == 0 {
		t.Fatal("перечень ролей пуст")
	}
	if all[0] != plan.RoleEntity {
		t.Errorf("первая роль %q, ожидалась %q", all[0], plan.RoleEntity)
	}
	for i := 1; i < len(all); i++ {
		if all[i] <= all[i-1] {
			t.Errorf("порядок ролей нарушен на позиции %d", i)
		}
	}
}

func TestLayerKindNamesAreDefined(t *testing.T) {
	for _, kind := range []plan.LayerKind{plan.LayerDomain, plan.LayerApp, plan.LayerInfra} {
		if !kind.IsKnown() || kind.String() == "" {
			t.Errorf("вид слоя %d не опознан или без имени", int(kind))
		}
	}
	if plan.LayerUnknown.IsKnown() {
		t.Error("нулевое значение вида слоя опознано")
	}
	if plan.LayerKind(100).IsKnown() {
		t.Error("вид слоя вне перечня опознан")
	}
}

func TestPlanCarriesThreeLevels(t *testing.T) {
	structure := plan.Plan{
		Modules: []plan.Module{{
			Layers: []plan.Layer{{
				Kind:  plan.LayerDomain,
				Roles: []plan.Role{plan.RoleEntity},
			}},
		}},
		ProjectRoles: []plan.Role{plan.RoleEntrypoint},
	}

	if len(structure.Modules) != 1 {
		t.Fatalf("модулей %d, ожидался 1", len(structure.Modules))
	}
	if structure.Modules[0].Layers[0].Kind != plan.LayerDomain {
		t.Error("вид слоя прочитан неверно")
	}
	if structure.Modules[0].Layers[0].Roles[0] != plan.RoleEntity {
		t.Error("роль слоя прочитана неверно")
	}
	if structure.ProjectRoles[0] != plan.RoleEntrypoint {
		t.Error("роль уровня проекта прочитана неверно")
	}
}
