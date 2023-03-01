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
		if ref == nil || ref.Ref != "" {
			continue
		}
		ext := ref.ExtendedTypeInfo
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

func (s *schemaNormalizer) processSchemaRef(ref *Schema) *Schema {
	if ref == nil || ref.Ref != "" {
		return ref
	}
	ext := ref.ExtendedTypeInfo
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
		for _, mediaType := range ref.Content {
			mediaType.Schema = s.processSchemaRef(mediaType.Schema)
		}
	}
	requestBody := op.RequestBody
	if requestBody != nil && requestBody.Ref == "" {
		content := requestBody.Content
		for _, mediaType := range content {
			mediaType.Schema = s.processSchemaRef(mediaType.Schema)
		}
	}
}

func (s *schemaNormalizer) process(ref *Schema, args []*Schema) *Schema {
	schemaRef := Unref(s.doc, ref)
	res := schemaRef.Clone()
	res.SpecializedFromGeneric = true
	schema := res
	ext := schema.ExtendedTypeInfo
	specificTypeKey := s.modelKey(ref.GetKey(), args)
	resRef := RefComponentSchemas(specificTypeKey)
	if ref.Ref != "" {
		_, exists := s.doc.Components.Schemas[specificTypeKey]
		if exists {
			return resRef
		}
		res.ExtendedTypeInfo = NewSpecificExtendType(ref, args...)
		s.doc.Components.Schemas[specificTypeKey] = res
	}

	if ext != nil {
		switch ext.Type {
		case ExtendedTypeSpecific:
			return s.process(ext.SpecificType.Type, s.mergeArgs(ext.SpecificType.Args, args))
		case ExtendedTypeParam:
			arg := args[ext.TypeParam.Index]
			if arg == nil {
				return nil
			}
			res = arg
		}
	}

	if schema.Items != nil {
		res.Items = s.process(res.Items, args)
	}
	if schema.AdditionalProperties != nil {
		res.AdditionalProperties = s.process(res.AdditionalProperties, args)
	}
	for key, property := range schema.Properties {
		schema.Properties[key] = s.process(property, args)
	}
	if ref.Ref != "" {
		return resRef
	}
	return res
}

func (s *schemaNormalizer) mergeArgs(args []*Schema, args2 []*Schema) []*Schema {
	res := make([]*Schema, 0, len(args))
	for _, _arg := range args {
		arg := _arg
		ext := arg.ExtendedTypeInfo
		if ext != nil && ext.Type == ExtendedTypeParam {
			arg = args2[ext.TypeParam.Index]
		}
		res = append(res, arg)
	}
	return res
}

func (s *schemaNormalizer) modelKey(key string, args []*Schema) string {
	sb := strings.Builder{}
	sb.WriteString(key)
	if len(args) <= 0 {
		return key
	}
	sb.WriteString("[" + args[0].GetKey())
	for _, ref := range args[1:] {
		sb.WriteString("," + ref.GetKey())
	}
	sb.WriteString("]")
	return sb.String()
}
