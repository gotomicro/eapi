package spec

import "strings"

type schemaNormalizer struct {
	doc *T
}

func newSchemaNormalizer(doc *T) *schemaNormalizer {
	return &schemaNormalizer{doc: doc}
}

func (s *schemaNormalizer) normalize() *T {
	for key, ref := range s.doc.Components.Schemas {
		if ref.Ref != "" || ref.Value == nil {
			continue
		}
		ext := ref.Value.ExtendedTypeInfo
		if ext == nil || ext.Type != ExtendedTypeSpecific {
			continue
		}

		s.doc.Components.Schemas[key] = s.process(ext.SpecificType.Type, ext.SpecificType.Args)
	}
	for _, item := range s.doc.Paths {
		s.processPathItem(item)
	}
	return s.doc
}

func (s *schemaNormalizer) processPathItem(item *PathItem) {
	s.processOperation(item.Connect)
	s.processOperation(item.Delete)
	s.processOperation(item.Get)
	s.processOperation(item.Head)
	s.processOperation(item.Options)
	s.processOperation(item.Patch)
	s.processOperation(item.Post)
	s.processOperation(item.Put)
	s.processOperation(item.Trace)
}

func (s *schemaNormalizer) processSchemaRef(ref *SchemaRef) *SchemaRef {
	if ref.Ref != "" || ref.Value == nil {
		return ref
	}
	ext := ref.Value.ExtendedTypeInfo
	if ext == nil || ext.Type != ExtendedTypeSpecific {
		return ref
	}
	return s.process(ext.SpecificType.Type, ext.SpecificType.Args)
}

func (s *schemaNormalizer) processOperation(op *Operation) {
	if op == nil {
		return
	}
	for _, ref := range op.Responses {
		for _, mediaType := range ref.Value.Content {
			mediaType.Schema = s.processSchemaRef(mediaType.Schema)
		}
	}
}

func (s *schemaNormalizer) process(ref *SchemaRef, args []*SchemaRef) *SchemaRef {
	schemaRef := Unref(s.doc, ref)
	res := schemaRef.Clone()
	res.Value.SpecializedFromGeneric = true
	schema := res.Value
	ext := schema.ExtendedTypeInfo
	specificTypeKey := s.modelKey(ref.Key(), args)
	refKey := "#/components/schemas/" + specificTypeKey
	if ref.Ref != "" {
		_, exists := s.doc.Components.Schemas[specificTypeKey]
		if exists {
			return NewSchemaRef(refKey, nil)
		}
		res.Value.ExtendedTypeInfo = NewSpecificExtendType(ref, args...)
		s.doc.Components.Schemas[specificTypeKey] = res
	}

	if ext != nil {
		switch ext.Type {
		case ExtendedTypeSpecific:
			return s.process(ext.SpecificType.Type, s.mergeArgs(ext.SpecificType.Args, args))
		case ExtendedTypeParam:
			res.Value = args[ext.TypeParam.Index].Value
		}
	}

	if schema.Items != nil {
		res.Value.Items = s.process(res.Value.Items, args)
	}
	if schema.AdditionalProperties != nil {
		res.Value.AdditionalProperties = s.process(res.Value.AdditionalProperties, args)
	}
	for key, property := range schema.Properties {
		schema.Properties[key] = s.process(property, args)
	}
	if ref.Ref != "" {
		return NewSchemaRef(refKey, nil)
	}
	return res
}

func (s *schemaNormalizer) mergeArgs(args []*SchemaRef, args2 []*SchemaRef) []*SchemaRef {
	res := make([]*SchemaRef, 0, len(args))
	for _, _arg := range args {
		arg := _arg
		ext := arg.Value.ExtendedTypeInfo
		if ext != nil && ext.Type == ExtendedTypeParam {
			arg = args2[ext.TypeParam.Index]
		}
		res = append(res, arg)
	}
	return res
}

func (s *schemaNormalizer) modelKey(key string, args []*SchemaRef) string {
	sb := strings.Builder{}
	sb.WriteString(key)
	if len(args) <= 0 {
		return key
	}
	sb.WriteString("[" + args[0].Key())
	for _, ref := range args[1:] {
		sb.WriteString("," + ref.Key())
	}
	sb.WriteString("]")
	return sb.String()
}
