package pongo2render

import (
	"strings"
	"unicode/utf8"

	"ego-gen-api/internal/parser"
	"ego-gen-api/internal/pongo2"
	"ego-gen-api/internal/utils"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-openapi/spec"
)

func init() {
	_ = pongo2.RegisterFilter("lowerFirst", pongo2LowerFirst)
	_ = pongo2.RegisterFilter("upperFirst", pongo2UpperFirst)
	_ = pongo2.RegisterFilter("snakeString", pongo2SnakeString)
	_ = pongo2.RegisterFilter("camelString", pongo2CamelString)
	_ = pongo2.RegisterFilter("getType", getType)
	_ = pongo2.RegisterFilter("getDefinitionName", getDefinitionName)
	_ = pongo2.RegisterFilter("getFieldTypescriptType", getFieldTypescriptType)
	_ = pongo2.RegisterFilter("getApiName", getApiName)
	_ = pongo2.RegisterFilter("getDescription", getDescription)

}

func pongo2LowerFirst(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if in.Len() <= 0 {
		return pongo2.AsSafeValue(""), nil
	}
	t := in.String()
	r, size := utf8.DecodeRuneInString(t)
	return pongo2.AsSafeValue(strings.ToLower(string(r)) + t[size:]), nil
}

func pongo2UpperFirst(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if in.Len() <= 0 {
		return pongo2.AsSafeValue(""), nil
	}
	t := in.String()
	return pongo2.AsSafeValue(strings.Replace(t, string(t[0]), strings.ToUpper(string(t[0])), 1)), nil
}

// snake string, XxYy to xx_yy
func pongo2SnakeString(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if in.Len() <= 0 {
		return pongo2.AsSafeValue(""), nil
	}
	t := in.String()
	return pongo2.AsSafeValue(utils.SnakeString(t)), nil
}

// snake string, XxYy to xx_yy
func pongo2CamelString(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if in.Len() <= 0 {
		return pongo2.AsSafeValue(""), nil
	}
	t := in.String()
	return pongo2.AsSafeValue(utils.CamelString(t)), nil
}

func upperFirst(str string) string {
	return strings.Replace(str, string(str[0]), strings.ToUpper(string(str[0])), 1)
}

func lowerFirst(str string) string {
	return strings.Replace(str, string(str[0]), strings.ToLower(string(str[0])), 1)
}

func getType(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if in.Len() <= 0 {
		return pongo2.AsSafeValue(""), nil
	}
	arr := []string(in.Interface().(spec.StringOrArray))
	return pongo2.AsSafeValue(arr[0]), nil
}

func getDefinitionName(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if in.Len() <= 0 {
		return pongo2.AsSafeValue(""), nil
	}
	return pongo2.AsSafeValue(getInnerDefinitionName(in.String())), nil
}

func getFieldTypescriptType(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	props := in.Interface().(spec.Schema)
	var str string
	str = getTsType(props)
	return pongo2.AsSafeValue(str), nil
}

func getTsType(props spec.Schema) string {
	var str string
	// 存在引用
	if props.Ref.String() != "" {
		str = getInnerDefinitionName(parser.GetSchemaDefinitionName(props.Ref.String()))
	} else {
		switch props.Type[0] {
		case parser.ARRAY:
			if props.Items.Schema.Ref.String() != "" {
				str = getInnerDefinitionName(parser.GetSchemaDefinitionName(props.Items.Schema.Ref.String())) + "[]"
			} else {

				switch props.Items.Schema.Type[0] {
				case parser.INTEGER:
					str = "number[]"
				case parser.STRING:
					str = "string[]"
				case parser.BOOLEAN:
					str = "boolean[]"
				}
			}
		case parser.OBJECT:
		case parser.BOOLEAN:
			str = "boolean"
		case parser.STRING:
			str = "string"
		case parser.INTEGER:
			str = "number"
		case parser.JSONRAW_MESSAGE:
			str = "any"
		}
	}
	if str == "" {
		spew.Dump(props)
	}
	return str
}

// shop.CreateReq => ShopCreateReq
func getInnerDefinitionName(req string) string {
	arr := strings.Split(req, ".")
	if len(arr) == 2 {
		return upperFirst(arr[0]) + upperFirst(arr[1])
	}
	return upperFirst(arr[0])
}

func getApiName(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {

	var str string
	value := in.Interface().(parser.UrlInfo)
	arr := strings.Split(value.FullPath, "/")
	// Get  /api/test/some
	first := upperFirst(strings.ToLower(value.Method))
	str = first
	for _, value := range arr {
		if value != "" {
			if strings.Contains(value, "-") {
				newArr := strings.Split(value, "-")
				for _, v := range newArr {
					if v == "" {
						continue
					}
					str += upperFirst(v)
				}
			} else if strings.Contains(value, ":") {
				newArr := strings.Split(value, ":")
				for _, v := range newArr {
					if v == "" {
						continue
					}
					str += upperFirst(v)
				}
			} else {
				str += upperFirst(value)
			}
		}
	}

	return pongo2.AsSafeValue(str), nil
}

func getDescription(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	var str string
	value := in.Interface().(spec.Schema)
	str = strings.Replace(value.Description, "\n", "\n//", -1)
	return pongo2.AsSafeValue(str), nil
}
