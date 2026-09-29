// Package app строит план структуры из измерений профиля.
//
// Измерение «язык» в этот слой не передаётся: построение принимает три измерения, а не профиль
// целиком, поэтому сослаться на язык здесь физически не на что. Критерий «ни один компонент фазы
// не ссылается на язык» проверяется компилятором, а не ревью.
package app

import (
	"fmt"

	plan "github.com/united-software-platform/extruder/internal/plan/domain"
	profile "github.com/united-software-platform/extruder/internal/profile/domain"
)

// UnsupportedValueError — значение измерения допущено профилем, но плану структуры неизвестно.
// Расхождение перечней профиля и плана обнаруживается явной ошибкой, а не молчаливо пустым планом.
type UnsupportedValueError struct {
	Dimension profile.Dimension
	Value     string
}

func (e UnsupportedValueError) Error() string {
	return fmt.Sprintf("измерение %q: значение %q планом структуры не поддержано",
		string(e.Dimension), e.Value)
}

// layerRole — роль, адресованная конкретному слою. Вкладчики, которые добавляют роли
// в существующие слои, возвращают срез таких пар, а не ассоциативный массив: порядок обхода
// массива в Go случаен, и эталон плана в тесте стал бы неоднозначным.
type layerRole struct {
	layer plan.LayerKind
	role  plan.Role
}

// BuildPlan строит план структуры из трёх измерений профиля.
func BuildPlan(
	projectType profile.ProjectType,
	architecture profile.Architecture,
	patterns profile.Patterns,
) (plan.Plan, error) {
	moduleCount, projectRoles, transportRoles, err := projectTypeContribution(projectType)
	if err != nil {
		return plan.Plan{}, err
	}

	layers, err := architectureContribution(architecture)
	if err != nil {
		return plan.Plan{}, err
	}

	patternRoles, err := patternsContribution(patterns)
	if err != nil {
		return plan.Plan{}, err
	}

	layers, err = withAddedRoles(layers, transportRoles)
	if err != nil {
		return plan.Plan{}, err
	}
	layers, err = withAddedRoles(layers, patternRoles)
	if err != nil {
		return plan.Plan{}, err
	}

	modules := make([]plan.Module, 0, moduleCount)
	for i := 0; i < moduleCount; i++ {
		modules = append(modules, plan.Module{Layers: cloneLayers(layers)})
	}

	return plan.Plan{Modules: modules, ProjectRoles: projectRoles}, nil
}

// withAddedRoles дописывает адресованные роли в конец соответствующих слоёв.
func withAddedRoles(layers []plan.Layer, added []layerRole) ([]plan.Layer, error) {
	result := cloneLayers(layers)
	for _, item := range added {
		index := indexOfLayer(result, item.layer)
		if index < 0 {
			return nil, fmt.Errorf("слоя %q в плане нет: роль %q добавить некуда",
				item.layer, item.role)
		}
		result[index].Roles = append(result[index].Roles, item.role)
	}
	return result, nil
}

// indexOfLayer возвращает позицию слоя указанного вида или -1.
func indexOfLayer(layers []plan.Layer, kind plan.LayerKind) int {
	for i, layer := range layers {
		if layer.Kind == kind {
			return i
		}
	}
	return -1
}

// cloneLayers копирует слои вместе с их ролями: модули не должны делить один и тот же срез.
func cloneLayers(layers []plan.Layer) []plan.Layer {
	result := make([]plan.Layer, 0, len(layers))
	for _, layer := range layers {
		roles := make([]plan.Role, len(layer.Roles))
		copy(roles, layer.Roles)
		result = append(result, plan.Layer{Kind: layer.Kind, Roles: roles})
	}
	return result
}
