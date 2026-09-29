package app

import (
	plan "github.com/united-software-platform/extruder/internal/plan/domain"
	profile "github.com/united-software-platform/extruder/internal/profile/domain"
)

// projectTypeContribution вносит вклад измерения «тип проекта»: сколько модулей в проекте,
// какие роли принадлежат проекту целиком и какая транспортная роль попадает в слой.
//
// Транспортная роль отнесена к типу проекта, а не к архитектуре намеренно: обработчик входящего
// запроса нужен сервису и не нужен библиотеке, тогда как слои у них одни и те же.
func projectTypeContribution(projectType profile.ProjectType) (int, []plan.Role, []layerRole, error) {
	switch projectType.String() {
	case "service":
		return 1,
			[]plan.Role{plan.RoleModuleManifest, plan.RoleEntrypoint, plan.RoleBuildTarget},
			[]layerRole{{layer: plan.LayerInfra, role: plan.RoleInboundHandler}},
			nil
	default:
		return 0, nil, nil, UnsupportedValueError{
			Dimension: profile.DimensionProjectType,
			Value:     projectType.String(),
		}
	}
}
