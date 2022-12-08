package annotation

type Type int

const (
	Required Type = iota + 1
	Consume
	Produce
	Unresolved
	Ignore
	Tag
	Description
	Summary
	ID
	Deprecated
)

type Annotation interface {
	Type() Type
}

type SimpleAnnotation struct {
	t Type
}

func (a *SimpleAnnotation) Type() Type {
	return a.t
}

func newSimpleAnnotation(t Type) *SimpleAnnotation {
	return &SimpleAnnotation{t: t}
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

type TagAnnotation struct {
	Tag string
}

func (a *TagAnnotation) Type() Type {
	return Tag
}

type DescriptionAnnotation struct {
	Text string
}

func (a *DescriptionAnnotation) Type() Type {
	return Description
}

type SummaryAnnotation struct {
	Text string
}

func (a *SummaryAnnotation) Type() Type {
	return Summary
}

type IdAnnotation struct {
	Text string
}

func (a *IdAnnotation) Type() Type {
	return ID
}
