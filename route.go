package analyzer

import (
	"go/ast"
	"net/http"

	"github.com/go-openapi/spec"
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
	*spec.Operation
}

func NewAPISpec(id string) *APISpec {
	return &APISpec{
		Operation: spec.NewOperation(id),
	}
}

// LoadFromFuncDecl load annotations/description from comments of handler function
func (s *APISpec) LoadFromFuncDecl(funcDecl *ast.FuncDecl) {
	cg := funcDecl.Doc
	comment := ParseComment(cg)
	if comment != nil {
		s.WithConsumes(comment.Consumes()...)
		s.WithProduces(comment.Produces()...)
		s.WithDescription(comment.TrimPrefix(funcDecl.Name.Name))
	}
}
