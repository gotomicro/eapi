package parser

import (
	"errors"
)

const (
	jsonTag             = "json"
	bindingTag          = "binding"
	defaultTag          = "default"
	enumsTag            = "enums"
	exampleTag          = "example"
	schemaExampleTag    = "schemaExample"
	formatTag           = "format"
	validateTag         = "validate"
	minimumTag          = "minimum"
	maximumTag          = "maximum"
	minLengthTag        = "minlength"
	maxLengthTag        = "maxlength"
	multipleOfTag       = "multipleOf"
	readOnlyTag         = "readonly"
	extensionsTag       = "extensions"
	collectionFormatTag = "collectionFormat"
)

var (
	// ErrFuncTypeField field type is func.
	ErrFuncTypeField = errors.New("field type is func")

	// ErrFailedConvertPrimitiveType Failed to convert for swag to interpretable type.
	ErrFailedConvertPrimitiveType = errors.New("swag property: failed convert primitive type")

	// ErrSkippedField .swaggo specifies field should be skipped
	ErrSkippedField = errors.New("field is skipped by global overrides")
)

const (
	// PascalCase indicates using PascalCase strategy for struct field.
	PascalCase = "pascalcase"

	// SnakeCase indicates using SnakeCase strategy for struct field.
	SnakeCase = "snakecase"

	descriptionAttr = "@description"
)
