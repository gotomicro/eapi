package spec

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/jsoninfo"
	"github.com/go-openapi/jsonpointer"
)

func foundUnresolvedRef(ref string) error {
	return fmt.Errorf("found unresolved ref: %q", ref)
}

func failedToResolveRefFragmentPart(value, what string) error {
	return fmt.Errorf("failed to resolve %q in fragment in URI: %q", what, value)
}

// Ref is specified by OpenAPI/Swagger 3.0 standard.
// See https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#referenceObject
type Ref struct {
	Ref string `json:"$ref" yaml:"$ref"`
}

// CallbackRef represents either a Callback or a $ref to a Callback.
// When serializing and both fields are set, Ref is preferred over Value.
type CallbackRef struct {
	Ref   string
	Value *Callback
}

var _ jsonpointer.JSONPointable = (*CallbackRef)(nil)

// MarshalYAML returns the YAML encoding of CallbackRef.
func (value *CallbackRef) MarshalYAML() (interface{}, error) {
	return marshalRefYAML(value.Ref, value.Value)
}

// MarshalJSON returns the JSON encoding of CallbackRef.
func (value *CallbackRef) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalRef(value.Ref, value.Value)
}

// UnmarshalJSON sets CallbackRef to a copy of data.
func (value *CallbackRef) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalRef(data, &value.Ref, &value.Value)
}

// Validate returns an error if CallbackRef does not comply with the OpenAPI spec.
func (value *CallbackRef) Validate(ctx context.Context) error {
	if v := value.Value; v != nil {
		return v.Validate(ctx)
	}
	return foundUnresolvedRef(value.Ref)
}

// JSONLookup implements github.com/go-openapi/jsonpointer#JSONPointable
func (value CallbackRef) JSONLookup(token string) (interface{}, error) {
	if token == "$ref" {
		return value.Ref, nil
	}

	ptr, _, err := jsonpointer.GetForToken(value.Value, token)
	return ptr, err
}

// ExampleRef represents either a Example or a $ref to a Example.
// When serializing and both fields are set, Ref is preferred over Value.
type ExampleRef struct {
	Ref   string
	Value *Example
}

var _ jsonpointer.JSONPointable = (*ExampleRef)(nil)

// MarshalYAML returns the YAML encoding of ExampleRef.
func (value *ExampleRef) MarshalYAML() (interface{}, error) {
	return marshalRefYAML(value.Ref, value.Value)
}

// MarshalJSON returns the JSON encoding of ExampleRef.
func (value *ExampleRef) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalRef(value.Ref, value.Value)
}

// UnmarshalJSON sets ExampleRef to a copy of data.
func (value *ExampleRef) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalRef(data, &value.Ref, &value.Value)
}

// Validate returns an error if ExampleRef does not comply with the OpenAPI spec.
func (value *ExampleRef) Validate(ctx context.Context) error {
	if v := value.Value; v != nil {
		return v.Validate(ctx)
	}
	return foundUnresolvedRef(value.Ref)
}

// JSONLookup implements github.com/go-openapi/jsonpointer#JSONPointable
func (value ExampleRef) JSONLookup(token string) (interface{}, error) {
	if token == "$ref" {
		return value.Ref, nil
	}

	ptr, _, err := jsonpointer.GetForToken(value.Value, token)
	return ptr, err
}

// HeaderRef represents either a Header or a $ref to a Header.
// When serializing and both fields are set, Ref is preferred over Value.
type HeaderRef struct {
	Ref   string
	Value *Header
}

var _ jsonpointer.JSONPointable = (*HeaderRef)(nil)

// MarshalYAML returns the YAML encoding of HeaderRef.
func (value *HeaderRef) MarshalYAML() (interface{}, error) {
	return marshalRefYAML(value.Ref, value.Value)
}

// MarshalJSON returns the JSON encoding of HeaderRef.
func (value *HeaderRef) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalRef(value.Ref, value.Value)
}

// UnmarshalJSON sets HeaderRef to a copy of data.
func (value *HeaderRef) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalRef(data, &value.Ref, &value.Value)
}

// Validate returns an error if HeaderRef does not comply with the OpenAPI spec.
func (value *HeaderRef) Validate(ctx context.Context) error {
	if v := value.Value; v != nil {
		return v.Validate(ctx)
	}
	return foundUnresolvedRef(value.Ref)
}

// JSONLookup implements github.com/go-openapi/jsonpointer#JSONPointable
func (value HeaderRef) JSONLookup(token string) (interface{}, error) {
	if token == "$ref" {
		return value.Ref, nil
	}

	ptr, _, err := jsonpointer.GetForToken(value.Value, token)
	return ptr, err
}

// LinkRef represents either a Link or a $ref to a Link.
// When serializing and both fields are set, Ref is preferred over Value.
type LinkRef struct {
	Ref   string
	Value *Link
}

// MarshalYAML returns the YAML encoding of LinkRef.
func (value *LinkRef) MarshalYAML() (interface{}, error) {
	return marshalRefYAML(value.Ref, value.Value)
}

// MarshalJSON returns the JSON encoding of LinkRef.
func (value *LinkRef) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalRef(value.Ref, value.Value)
}

// UnmarshalJSON sets LinkRef to a copy of data.
func (value *LinkRef) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalRef(data, &value.Ref, &value.Value)
}

// Validate returns an error if LinkRef does not comply with the OpenAPI spec.
func (value *LinkRef) Validate(ctx context.Context) error {
	if v := value.Value; v != nil {
		return v.Validate(ctx)
	}
	return foundUnresolvedRef(value.Ref)
}

// ParameterRef represents either a Parameter or a $ref to a Parameter.
// When serializing and both fields are set, Ref is preferred over Value.
type ParameterRef = Parameter

// ResponseRef represents either a Response or a $ref to a Response.
// When serializing and both fields are set, Ref is preferred over Value.
type ResponseRef = Response

// RequestBodyRef represents either a RequestBody or a $ref to a RequestBody.
// When serializing and both fields are set, Ref is preferred over Value.
type RequestBodyRef = RequestBody

// SchemaRef represents either a Schema or a $ref to a Schema.
// When serializing and both fields are set, Ref is preferred over Value.
type SchemaRef = Schema

type schemaRef struct {
	Summary     string `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Ref         string `json:"$ref,omitempty" yaml:"$ref,omitempty"`
}

func (value *Schema) Unref(doc *T) *Schema {
	if value.Ref != "" {
		return doc.GetSchemaByRef(value.Ref)
	}
	return value
}

func (value *Schema) GetKey() string {
	if value == nil {
		return ""
	}
	if value.Ref != "" {
		return strings.TrimPrefix(value.Ref, "#/components/schemas/")
	}
	if value.Key != "" {
		return value.Key
	}
	if value.Title != "" {
		return value.Title
	}
	switch value.Type {
	case TypeBoolean, TypeInteger, TypeNumber, TypeString:
		return value.Type

	default: // TypeArray, TypeObject:
		h := md5.New()
		b, _ := json.Marshal(value)
		h.Write(b)
		md5Hex := fmt.Sprintf("%x", h.Sum(nil))
		return md5Hex[:12]
	}
}

func (value *Schema) IsTypeAlias() bool {
	if value.Ref != "" {
		return true
	}
	ext := value.ExtendedTypeInfo
	if ext == nil {
		return false
	}
	return ext.Type == ExtendedTypeParam && len(ext.TypeParams) == 1
}

// SecuritySchemeRef represents either a SecurityScheme or a $ref to a SecurityScheme.
// When serializing and both fields are set, Ref is preferred over Value.
type SecuritySchemeRef struct {
	Ref   string
	Value *SecurityScheme
}

var _ jsonpointer.JSONPointable = (*SecuritySchemeRef)(nil)

// MarshalYAML returns the YAML encoding of SecuritySchemeRef.
func (value *SecuritySchemeRef) MarshalYAML() (interface{}, error) {
	return marshalRefYAML(value.Ref, value.Value)
}

// MarshalJSON returns the JSON encoding of SecuritySchemeRef.
func (value *SecuritySchemeRef) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalRef(value.Ref, value.Value)
}

// UnmarshalJSON sets SecuritySchemeRef to a copy of data.
func (value *SecuritySchemeRef) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalRef(data, &value.Ref, &value.Value)
}

// Validate returns an error if SecuritySchemeRef does not comply with the OpenAPI spec.
func (value *SecuritySchemeRef) Validate(ctx context.Context) error {
	if v := value.Value; v != nil {
		return v.Validate(ctx)
	}
	return foundUnresolvedRef(value.Ref)
}

// JSONLookup implements github.com/go-openapi/jsonpointer#JSONPointable
func (value SecuritySchemeRef) JSONLookup(token string) (interface{}, error) {
	if token == "$ref" {
		return value.Ref, nil
	}

	ptr, _, err := jsonpointer.GetForToken(value.Value, token)
	return ptr, err
}

// marshalRefYAML returns the YAML encoding of ref values.
func marshalRefYAML(value string, otherwise interface{}) (interface{}, error) {
	if value != "" {
		return &Ref{Ref: value}, nil
	}
	return otherwise, nil
}
