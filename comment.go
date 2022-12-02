package eapi

import (
	"go/ast"
	"strings"

	"github.com/gotomicro/eapi/annotation"
	"github.com/gotomicro/eapi/spec"
	"github.com/samber/lo"
)

type Comment struct {
	Text        string
	Annotations []annotation.Annotation
}

// TrimPrefix trim comment prefix and return trimmed string
func (c *Comment) TrimPrefix(prefix string) string {
	return strings.TrimPrefix(c.Text, prefix)
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

func (c *Comment) Nullable() bool {
	for _, a := range c.Annotations {
		if a.Type() == annotation.Nullable {
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
		schema.Description = c.Text
		return
	}

	value := schema.Value
	if value == nil {
		return
	}
	value.Description = c.Text
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
		_, ok := annot.(*annotation.IgnoreAnnotation)
		if ok {
			return true
		}
	}
	return false
}

func (c *Comment) Summary() string {
	for _, annot := range c.Annotations {
		summary, ok := annot.(*annotation.SummaryAnnotation)
		if ok {
			return strings.TrimSpace(summary.Text)
		}
	}
	return ""
}

func (c *Comment) ID() string {
	for _, annot := range c.Annotations {
		id, ok := annot.(*annotation.IdAnnotation)
		if ok {
			return strings.TrimSpace(id.Text)
		}
	}
	return ""
}

func ParseComment(commentGroup *ast.CommentGroup) *Comment {
	if commentGroup == nil {
		return nil
	}
	c := &Comment{}
	var lines []string
	for _, comment := range commentGroup.List {
		line := strings.TrimPrefix(comment.Text, "//")
		annot := annotation.NewParser(line).Parse()
		if annot != nil {
			c.Annotations = append(c.Annotations, annot)
			desc, ok := annot.(*annotation.DescriptionAnnotation)
			if ok {
				lines = append(lines, strings.TrimSpace(desc.Text))
			}
		} else {
			lines = append(lines, strings.TrimSpace(line))
		}
	}
	c.Text = strings.Join(lines, "\n\n")
	return c
}
