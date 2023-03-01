package spec

type ExtendedType = string

const (
	ExtendedTypeMap      ExtendedType = "map"
	ExtendedTypeAny      ExtendedType = "any"
	ExtendedTypeEnum     ExtendedType = "enum"
	ExtendedTypeSpecific ExtendedType = "specific"
	ExtendedTypeParam    ExtendedType = "param"
	ExtendedTypeObject   ExtendedType = "object"
	ExtendedTypeArray    ExtendedType = "array"
	ExtendedTypeNull     ExtendedType = "null"
	ExtendedTypeUnknown  ExtendedType = "unknown"
)

type ExtendedTypeInfo struct {
	Type string `json:"type,omitempty"`

	// for array
	Items *Schema `json:"items,omitempty"`

	MapKey   *Schema `json:"mapKey,omitempty"`
	MapValue *Schema `json:"mapValue,omitempty"`

	EnumItems []*ExtendedEnumItem `json:"enumItems,omitempty"`

	SpecificType *SpecificType `json:"specificType,omitempty"`

	// for generic type params
	TypeParams []*TypeParam `json:"typeParams,omitempty"`

	TypeParam *TypeParam `json:"typeParam,omitempty"`
}

func NewExtendedEnumType(items ...*ExtendedEnumItem) *ExtendedTypeInfo {
	return &ExtendedTypeInfo{
		Type:      ExtendedTypeEnum,
		MapValue:  nil,
		EnumItems: items,
	}
}

func NewAnyExtendedType() *ExtendedTypeInfo {
	return &ExtendedTypeInfo{Type: ExtendedTypeAny}
}

func NewMapExtendedType(key, value *Schema) *ExtendedTypeInfo {
	return &ExtendedTypeInfo{
		Type:     ExtendedTypeMap,
		MapKey:   key,
		MapValue: value,
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

type SpecificType struct {
	Args []*Schema `json:"args"`
	Type *Schema   `json:"type"`
}

func NewSpecificExtendType(genericType *Schema, args ...*Schema) *ExtendedTypeInfo {
	return &ExtendedTypeInfo{
		Type:         ExtendedTypeSpecific,
		SpecificType: &SpecificType{Args: args, Type: genericType},
	}
}

type TypeParam struct {
	Index      int    `json:"index"`
	Name       string `json:"name"`
	Constraint string `json:"constraint"`
}

func NewTypeParamExtendedType(param *TypeParam) *ExtendedTypeInfo {
	return &ExtendedTypeInfo{Type: ExtendedTypeParam, TypeParam: param}
}

func NewObjectExtType() *ExtendedTypeInfo {
	return &ExtendedTypeInfo{Type: ExtendedTypeObject}
}

func NewArrayExtType(items *Schema) *ExtendedTypeInfo {
	return &ExtendedTypeInfo{Type: ExtendedTypeArray, Items: items}
}

func NewNullExtType() *ExtendedTypeInfo {
	return &ExtendedTypeInfo{Type: ExtendedTypeNull}
}

func NewUnknownExtType() *ExtendedTypeInfo {
	return &ExtendedTypeInfo{Type: ExtendedTypeUnknown}
}
