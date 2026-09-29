// Package domain содержит контракт между фазами преобразования: план структуры в понятиях
// модулей, слоёв и ролей. О целевом языке этот слой не знает ничего — ни имён файлов,
// ни расширений, ни содержимого.
package domain

// Role — назначение отдельного элемента: внутри слоя или в проекте целиком.
// Перечень закрыт: роль, которой в нём нет, в план не попадает.
type Role int

// Роли, поддерживаемые этой версией. Порядок объявления задаёт порядок в плане,
// поэтому дописывать новую роль следует в конец, а не в середину.
const (
	RoleUnknown Role = iota
	RoleEntity
	RoleRepositoryContract
	RoleUseCase
	RoleRepositoryImpl
	RoleInboundHandler
	RoleModuleManifest
	RoleEntrypoint
	RoleBuildTarget
)

// roleNames — единственная точка определения имён ролей: отсюда их берут и печать,
// и проверка принадлежности перечню.
var roleNames = []string{
	RoleUnknown:            "",
	RoleEntity:             "entity",
	RoleRepositoryContract: "repository-contract",
	RoleUseCase:            "use-case",
	RoleRepositoryImpl:     "repository-impl",
	RoleInboundHandler:     "inbound-handler",
	RoleModuleManifest:     "module-manifest",
	RoleEntrypoint:         "entrypoint",
	RoleBuildTarget:        "build-target",
}

// String возвращает имя роли. У роли вне перечня имени нет.
func (r Role) String() string {
	if !r.IsKnown() {
		return ""
	}
	return roleNames[r]
}

// IsKnown сообщает, принадлежит ли роль перечню этой версии.
func (r Role) IsKnown() bool {
	return r > RoleUnknown && int(r) < len(roleNames)
}

// Roles возвращает все роли перечня в порядке объявления.
func Roles() []Role {
	all := make([]Role, 0, len(roleNames)-1)
	for role := RoleEntity; int(role) < len(roleNames); role++ {
		all = append(all, role)
	}
	return all
}
