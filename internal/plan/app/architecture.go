package app

import (
	plan "github.com/united-software-platform/extruder/internal/plan/domain"
	profile "github.com/united-software-platform/extruder/internal/profile/domain"
)

// architectureContribution вносит вклад измерения «архитектура»: слои модуля и базовые роли
// внутри них. Порядок слоёв и ролей задаётся здесь объявлением и далее не переставляется.
func architectureContribution(architecture profile.Architecture) ([]plan.Layer, error) {
	switch architecture.String() {
	case "layered":
		return []plan.Layer{
			{Kind: plan.LayerDomain, Roles: []plan.Role{
				plan.RoleEntity,
				plan.RoleRepositoryContract,
			}},
			{Kind: plan.LayerApp, Roles: []plan.Role{
				plan.RoleUseCase,
			}},
			{Kind: plan.LayerInfra, Roles: []plan.Role{
				plan.RoleRepositoryImpl,
			}},
		}, nil
	default:
		return nil, UnsupportedValueError{
			Dimension: profile.DimensionArchitecture,
			Value:     architecture.String(),
		}
	}
}
