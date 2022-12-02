package spec

type ExtendedType string

const (
	ExtendedTypeMap  ExtendedType = "map"
	ExtendedTypeAny  ExtendedType = "any"
	ExtendedTypeEnum ExtendedType = "enum"
)

type ExtendedTypeInfo struct {
	Type ExtendedType `json:"type"`

	// When Type = 'map'. Value means type of value of map.
	Value *Schema `json:"valueType"`

	// Enum Items
	EnumItems []*ExtendedEnumItem `json:"enumItems"`
}

type ExtendedEnumItem struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

func NewExtendedEnumType(items ...*ExtendedEnumItem) *ExtendedTypeInfo {
	return &ExtendedTypeInfo{
		Type:      ExtendedTypeEnum,
		Value:     nil,
		EnumItems: items,
	}
}

func NewAnyExtendedType() *ExtendedTypeInfo {
	return &ExtendedTypeInfo{Type: ExtendedTypeAny}
}
