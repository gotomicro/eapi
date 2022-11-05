package analyzer

import (
	"go/types"
	"net/http"

	"github.com/go-openapi/spec"
)

type RouteGroup struct {
	Prefix string
}

type Route struct {
	Method   string
	FullPath string
	Handler  *types.Func

	spec *APISpec
}

func (r *Route) Spec() *APISpec {
	return r.spec
}

func (r *Route) applyToPathItem(pathItem *spec.PathItem) {
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

func (r *Route) Operation() *spec.Operation {
	return r.spec.Operation
}

type APISpec struct {
	// TODO: 数据格式需要方便修改
	*spec.Operation
}
