package annotation

type Type int

const (
	Required Type = iota + 1
	Nullable
	Consume
	Produce
	Unresolved
	Ignore
)

type Annotation interface {
	Type() Type
}

type RequiredAnnotation struct{}

func (a *RequiredAnnotation) Type() Type {
	return Required
}

type NullableAnnotation struct{}

func (a *NullableAnnotation) Type() Type {
	return Nullable
}

type ConsumeAnnotation struct {
	ContentType string
}

func (a *ConsumeAnnotation) Type() Type {
	return Consume
}

type ProduceAnnotation struct {
	ContentType string
}

func (a *ProduceAnnotation) Type() Type {
	return Produce
}

type UnresolvedAnnotation struct {
	Tag    string
	Tokens []*Token
}

func (a *UnresolvedAnnotation) Type() Type {
	return Unresolved
}

type IgnoreAnnotation struct{}

func (a *IgnoreAnnotation) Type() Type {
	return Ignore
}
