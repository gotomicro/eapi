package analyzer

import (
	"go/ast"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
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
	return &API{Method: method, FullPath: fullPath, Spec: NewAPISpec(method + "." + fullPath)}
}

func (r *API) applyToPathItem(pathItem *openapi3.PathItem) {
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

func (r *API) Operation() *openapi3.Operation {
	return r.Spec.Operation
}

type APIs []*API

func (r *APIs) add(items ...*API) {
	*r = append(*r, items...)
}

type APISpec struct {
	Consumes []string
	*openapi3.Operation
}

func NewAPISpec(id string) *APISpec {
	op := openapi3.NewOperation()
	op.OperationID = id
	return &APISpec{
		Operation: op,
	}
}

// LoadFromFuncDecl load annotations/description from comments of handler function
func (s *APISpec) LoadFromFuncDecl(funcDecl *ast.FuncDecl) {
	cg := funcDecl.Doc
	comment := ParseComment(cg)
	if comment != nil {
		s.Summary = comment.Summary()
		s.Description = strings.TrimSpace(comment.TrimPrefix(funcDecl.Name.Name))
		if s.Summary == "" {
			s.Summary = s.Description
		}
		s.Tags = comment.Tags()
		s.OperationID = comment.ID()
		s.Consumes = append(s.Consumes, comment.Consumes()...)
	}
}
