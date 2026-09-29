// Package infra содержит предъявление плана структуры пользователю.
package infra

import (
	"fmt"
	"io"

	plan "github.com/united-software-platform/extruder/internal/plan/domain"
)

// Символы ветвления дерева. Разделителя пути среди них нет намеренно: план не содержит путей,
// и печать не должна вносить их видимость.
const (
	branch     = "├── "
	lastBranch = "└── "
	trunk      = "│   "
	blank      = "    "
)

// Write печатает план структуры деревом. Порядок узлов — тот, в котором их сложило построение:
// печать ничего не сортирует и не переставляет.
func Write(out io.Writer, structure plan.Plan) error {
	if _, err := fmt.Fprintln(out, "План структуры:"); err != nil {
		return err
	}
	for _, module := range structure.Modules {
		if err := writeModule(out, module); err != nil {
			return err
		}
	}
	// Роли уровня проекта идут на верхнем уровне, вне слоёв: они не принадлежат ни одному из них
	for _, role := range structure.ProjectRoles {
		if _, err := fmt.Fprintf(out, "роль %s\n", role); err != nil {
			return err
		}
	}
	return nil
}

// writeModule печатает модуль с его слоями и ролями.
func writeModule(out io.Writer, module plan.Module) error {
	if _, err := fmt.Fprintln(out, "модуль"); err != nil {
		return err
	}
	for i, layer := range module.Layers {
		layerPrefix, rolePrefix := branch, trunk
		if i == len(module.Layers)-1 {
			layerPrefix, rolePrefix = lastBranch, blank
		}
		if _, err := fmt.Fprintf(out, "%sслой %s\n", layerPrefix, layer.Kind); err != nil {
			return err
		}
		if err := writeRoles(out, layer.Roles, rolePrefix); err != nil {
			return err
		}
	}
	return nil
}

// writeRoles печатает роли слоя с заданным отступом.
func writeRoles(out io.Writer, roles []plan.Role, indent string) error {
	for i, role := range roles {
		prefix := branch
		if i == len(roles)-1 {
			prefix = lastBranch
		}
		if _, err := fmt.Fprintf(out, "%s%sроль %s\n", indent, prefix, role); err != nil {
			return err
		}
	}
	return nil
}
