package eapi

import (
	"github.com/gotomicro/eapi/annotation"
	"github.com/samber/lo"
)

type Environment struct {
	parent  *Environment
	records map[interface{}]interface{}
}

func NewEnvironment(parent *Environment) *Environment {
	return &Environment{parent: parent, records: make(map[interface{}]interface{})}
}

func (e *Environment) Define(k, v interface{}) *Environment {
	e.records[k] = v
	return e
}

func (e *Environment) Lookup(k interface{}) interface{} {
	env := e.Resolve(k)
	if env == nil {
		return nil
	}
	return env.records[k]
}

func (e *Environment) Resolve(k interface{}) *Environment {
	_, ok := e.records[k]
	if ok {
		return e
	}
	if e.parent != nil {
		return e.parent.Resolve(k)
	}

	return nil
}

func (e *Environment) Assign(k, v interface{}) *Environment {
	scope := e.Resolve(k)
	if scope != nil {
		scope.Define(k, v)
	}
	return e
}

type CommentStack struct {
	parent  *CommentStack
	comment *Comment
}

func NewCommentStack(parent *CommentStack, comment *Comment) *CommentStack {
	return &CommentStack{parent: parent, comment: comment}
}

func (e *CommentStack) ResolveByAnnotation(annotType annotation.Type) *CommentStack {
	if e.comment != nil {
		for _, a := range e.comment.Annotations {
			if a.Type() == annotType {
				return e
			}
		}
	}
	if e.parent != nil {
		return e.parent.ResolveByAnnotation(annotType)
	}
	return nil
}

func (e *CommentStack) LookupAnnotations(annotType annotation.Type) []annotation.Annotation {
	env := e.ResolveByAnnotation(annotType)
	if env == nil {
		return nil
	}
	return lo.Filter(env.comment.Annotations, func(a annotation.Annotation, i int) bool {
		return a.Type() == annotType
	})
}

func (e *CommentStack) LookupTags() []string {
	return lo.Map(e.LookupAnnotations(annotation.Tag), func(t annotation.Annotation, i int) string {
		return t.(*annotation.TagAnnotation).Tag
	})
}
