package domain

// LayerKind — вид слоя модуля. Перечень закрыт по той же причине, что и перечень ролей:
// вид слоя вне перечня в план не попадает.
type LayerKind int

// Виды слоёв, поддерживаемые этой версией.
const (
	LayerUnknown LayerKind = iota
	LayerDomain
	LayerApp
	LayerInfra
)

// layerKindNames — единственная точка определения имён видов слоёв.
var layerKindNames = []string{
	LayerUnknown: "",
	LayerDomain:  "domain",
	LayerApp:     "app",
	LayerInfra:   "infra",
}

// String возвращает имя вида слоя. У вида вне перечня имени нет.
func (k LayerKind) String() string {
	if !k.IsKnown() {
		return ""
	}
	return layerKindNames[k]
}

// IsKnown сообщает, принадлежит ли вид слоя перечню этой версии.
func (k LayerKind) IsKnown() bool {
	return k > LayerUnknown && int(k) < len(layerKindNames)
}
