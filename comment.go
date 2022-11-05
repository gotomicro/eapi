package analyzer

import (
	"go/ast"
	"strings"

	"ego-gen-api/annotation"

	"github.com/go-openapi/spec"
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

func (c *Comment) transformIntoSchema(schema *spec.Schema) {
	if schema == nil {
		return
	}
	schema.Description = c.Text
	schema.Nullable = c.Nullable()
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
	return res
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
		} else {
			lines = append(lines, strings.TrimSpace(line))
		}
	}
	c.Text = strings.Join(lines, "\n")
	return c
}
