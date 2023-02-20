package eapi

import (
	"go/ast"
	"net/http"
	"strings"

	"github.com/gotomicro/eapi/annotation"
	"github.com/gotomicro/eapi/spec"
)

type RouteGroup struct {
	Prefix string
}

type API struct {
	Method   string
	FullPath string
	Spec     *APISpec
}

func NewAPI(method string, fullPath string) *API {
	return &API{
		Method:   method,
		FullPath: fullPath,
		Spec:     NewAPISpec(),
	}
}

func (r *API) applyToPathItem(pathItem *spec.PathItem) {
	switch r.Method {
	case http.MethodGet:
		pathItem.Get = r.Operation()
	case http.MethodHead:
		pathItem.Head = r.Operation()
	case http.MethodPost:
		pathItem.Post = r.Operation()
	case http.MethodPut:
		pathItem.Put = r.Operation()
	case http.MethodPatch:
		pathItem.Patch = r.Operation()
	case http.MethodDelete:
		pathItem.Delete = r.Operation()
	case http.MethodOptions:
		pathItem.Options = r.Operation()
	case http.MethodTrace:
		pathItem.Trace = r.Operation()
	}
}

func (r *API) Operation() *spec.Operation {
	return r.Spec.Operation
}

type APIs []*API

func (r *APIs) add(items ...*API) {
	*r = append(*r, items...)
}

type APISpec struct {
	Consumes []string
	*spec.Operation
}

func NewAPISpec() *APISpec {
	op := spec.NewOperation()
	op.Responses = spec.NewResponses()
	delete(op.Responses, "default")

	return &APISpec{
		Operation: op,
	}
}

// LoadFromFuncDecl load annotations/description from comments of handler function
func (s *APISpec) LoadFromFuncDecl(ctx *Context, funcDecl *ast.FuncDecl) {
	cg := funcDecl.Doc
	comment := ParseComment(cg, ctx.Package().Fset)
	s.LoadFromComment(ctx, comment)
	if s.Description == "" {
		// 使用注释里的普通文本作为描述
		s.Description = strings.TrimSpace(strings.TrimPrefix(comment.Text(), funcDecl.Name.Name))
	}
}

func (s *APISpec) LoadFromComment(ctx *Context, comment *Comment) {
	if comment != nil {
		if s.Description == "" {
			s.Description = strings.Join(comment.Description(), "\n\n")
		}
		if s.Summary == "" {
			s.Summary = comment.Summary()
		}
		if len(s.Tags) == 0 {
			s.Tags = comment.Tags()
		}
		if s.OperationID == "" {
			s.OperationID = comment.ID()
		}
		if len(s.Consumes) == 0 {
			s.Consumes = append(s.Consumes, comment.Consumes()...)
		}
		if !s.Deprecated {
			comment.Deprecated()
		}
		if s.Security == nil {
			s.Security = comment.Security()
		}
	}
	if len(s.Tags) == 0 {
		s.Tags = ctx.Env.LookupTags()
	}
	if s.Security == nil {
		s.Security = convertSecAnnotationToSecurityRequirements(ctx.Env.LookupAnnotations(annotation.Security))
	}
}
