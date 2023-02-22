package eapi

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"strings"

	"github.com/gotomicro/eapi/annotation"
	"github.com/gotomicro/eapi/spec"
	"github.com/samber/lo"
)

type Comment struct {
	text        string
	Annotations []annotation.Annotation
}

func (c *Comment) Text() string {
	if c == nil {
		return ""
	}
	return c.text
}

func (c *Comment) TextPointer() *string {
	if c == nil || c.Text() == "" {
		return nil
	}
	return &c.text
}

// TrimPrefix trim comment prefix and return trimmed string
func (c *Comment) TrimPrefix(prefix string) string {
	if c == nil {
		return ""
	}
	return strings.TrimPrefix(c.text, prefix)
}

func (c *Comment) Required() bool {
	if c == nil {
		return false
	}

	for _, a := range c.Annotations {
		if a.Type() == annotation.Required {
			return true
		}
	}
	return false
}

func (c *Comment) Deprecated() bool {
	if c == nil {
		return false
	}
	for _, a := range c.Annotations {
		if a.Type() == annotation.Deprecated {
			return true
		}
	}
	return false
}

func (c *Comment) ApplyToSchema(schema *spec.SchemaRef) {
	if c == nil || schema == nil {
		return
	}
	if schema.Ref != "" {
		schema.Description = c.Text()
		schema.Summary = c.Summary()
		return
	}

	value := schema.Value
	if value == nil {
		return
	}
	value.Description = c.Text()
	value.Deprecated = c.Deprecated()
}

func (c *Comment) Consumes() []string {
	var res []string
	for _, annot := range c.Annotations {
		consume, ok := annot.(*annotation.ConsumeAnnotation)
		if ok {
			res = append(res, consume.ContentType)
		}
	}
	return res
}

func (c *Comment) Produces() []string {
	var res []string
	for _, annot := range c.Annotations {
		produce, ok := annot.(*annotation.ProduceAnnotation)
		if ok {
			res = append(res, produce.ContentType)
		}
	}
	return lo.Uniq(res)
}

func (c *Comment) Tags() []string {
	if c == nil {
		return nil
	}

	var res []string
	for _, annot := range c.Annotations {
		tags, ok := annot.(*annotation.TagAnnotation)
		if ok {
			res = append(res, tags.Tag)
		}
	}
	return res
}

func (c *Comment) Ignore() bool {
	if c == nil {
		return false
	}
	for _, annot := range c.Annotations {
		if annot.Type() == annotation.Ignore {
			return true
		}
	}
	return false
}

func (c *Comment) Summary() string {
	if c == nil {
		return ""
	}
	for _, annot := range c.Annotations {
		summary, ok := annot.(*annotation.SummaryAnnotation)
		if ok {
			return strings.TrimSpace(summary.Text)
		}
	}
	return ""
}

func (c *Comment) ID() string {
	if c == nil {
		return ""
	}
	for _, annot := range c.Annotations {
		id, ok := annot.(*annotation.IdAnnotation)
		if ok {
			return strings.TrimSpace(id.Text)
		}
	}
	return ""
}

func (c *Comment) Security() *spec.SecurityRequirements {
	if c == nil {
		return nil
	}

	return convertSecAnnotationToSecurityRequirements(c.Annotations)
}

func convertSecAnnotationToSecurityRequirements(annotations []annotation.Annotation) *spec.SecurityRequirements {
	ret := spec.NewSecurityRequirements()
	for _, annot := range annotations {
		annot, ok := annot.(*annotation.SecurityAnnotation)
		if ok {
			ret.With(spec.NewSecurityRequirement().Authenticate(annot.Name, annot.Params...))
		}
	}
	if len(*ret) == 0 {
		return nil
	}

	return ret
}

func ParseComment(commentGroup *ast.CommentGroup, fSet *token.FileSet) *Comment {
	if commentGroup == nil {
		return nil
	}
	c := &Comment{}
	var lines []string
	var descriptions []*annotation.DescriptionAnnotation
	for _, comment := range commentGroup.List {
		annot, err := annotation.NewParser(comment.Text).Parse()
		if err != nil {
			err := err.(*annotation.ParseError)
			fmt.Fprintf(os.Stderr, "[Invalid Annotation]: %s at %s\n", err, fSet.Position(comment.Pos()+token.Pos(err.Column)).String())
			continue
		}
		if annot != nil {
			c.Annotations = append(c.Annotations, annot)
			desc, ok := annot.(*annotation.DescriptionAnnotation)
			if ok {
				descriptions = append(descriptions, desc)
			}
		} else {
			line := strings.TrimPrefix(comment.Text, "//")
			lines = append(lines, strings.TrimSpace(line))
		}
	}

	if len(descriptions) > 0 {
		// 如果写了 @description 注解，则忽略普通注释文本
		lines = lo.Map(descriptions, func(t *annotation.DescriptionAnnotation, i int) string { return t.Text })
	}
	c.text = strings.Join(lines, "\n\n")
	return c
}
