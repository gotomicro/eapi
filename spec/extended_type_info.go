package spec

type ExtendedType string

const (
	ExtendedTypeMap  ExtendedType = "map"
	ExtendedTypeAny  ExtendedType = "any"
	ExtendedTypeEnum ExtendedType = "enum"
)

type ExtendedTypeInfo struct {
	Type ExtendedType `json:"type"`

	// When Type = 'map'. Key means type of key of map.
	Key *SchemaRef
	// When Type = 'map'. Value means type of value of map.
	Value *SchemaRef `json:"valueType"`

	// Enum Items
	EnumItems []*ExtendedEnumItem `json:"enumItems"`
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

func NewMapExtendedType(key, value *SchemaRef) *ExtendedTypeInfo {
	return &ExtendedTypeInfo{
		Type:  ExtendedTypeMap,
		Key:   key,
		Value: value,
	}
}

type ExtendedEnumItem struct {
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	Description string      `json:"description"`
}

func NewExtendEnumItem(key string, value interface{}, description string) *ExtendedEnumItem {
	return &ExtendedEnumItem{
		Key:         key,
		Value:       value,
		Description: description,
	}
}
