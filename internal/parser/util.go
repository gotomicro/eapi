package parser

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/gotomicro/ego/core/elog"
)

// defineTypeOfExample example value define the type (object and array unsupported)
func defineTypeOfExample(schemaType, arrayType, exampleValue string) (interface{}, error) {
	switch schemaType {
	case STRING:
		return exampleValue, nil
	case NUMBER:
		v, err := strconv.ParseFloat(exampleValue, 64)
		if err != nil {
			return nil, fmt.Errorf("example value %s can't convert to %s err: %s", exampleValue, schemaType, err)
		}

		return v, nil
	case INTEGER:
		v, err := strconv.Atoi(exampleValue)
		if err != nil {
			return nil, fmt.Errorf("example value %s can't convert to %s err: %s", exampleValue, schemaType, err)
		}

		return v, nil
	case BOOLEAN:
		v, err := strconv.ParseBool(exampleValue)
		if err != nil {
			return nil, fmt.Errorf("example value %s can't convert to %s err: %s", exampleValue, schemaType, err)
		}

		return v, nil
	case ARRAY:
		values := strings.Split(exampleValue, ",")
		result := make([]interface{}, 0)
		for _, value := range values {
			v, err := defineTypeOfExample(arrayType, "", value)
			if err != nil {
				return nil, err
			}
			result = append(result, v)
		}

		return result, nil
	case OBJECT:
		if arrayType == "" {
			return nil, fmt.Errorf("%s is unsupported type in example value `%s`", schemaType, exampleValue)
		}

		values := strings.Split(exampleValue, ",")
		result := map[string]interface{}{}
		for _, value := range values {
			mapData := strings.Split(value, ":")

			if len(mapData) == 2 {
				v, err := defineTypeOfExample(arrayType, "", mapData[1])
				if err != nil {
					return nil, err
				}
				result[mapData[0]] = v
			} else {
				return nil, fmt.Errorf("example value %s should format: key:value", exampleValue)
			}
		}

		return result, nil
	}

	return nil, fmt.Errorf("%s is unsupported type in example value %s", schemaType, exampleValue)
}

// defineType enum value define the type (object and array unsupported).
func defineType(schemaType string, value string) (v interface{}, err error) {
	schemaType = TransToValidSchemeType(schemaType)
	switch schemaType {
	case STRING:
		return value, nil
	case NUMBER:
		v, err = strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("enum value %s can't convert to %s err: %s", value, schemaType, err)
		}
	case INTEGER:
		v, err = strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("enum value %s can't convert to %s err: %s", value, schemaType, err)
		}
	case BOOLEAN:
		v, err = strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("enum value %s can't convert to %s err: %s", value, schemaType, err)
		}
	default:
		return nil, fmt.Errorf("%s is unsupported type in enum value %s", schemaType, value)
	}

	return v, nil
}

func getPackagePath(projectPath string) (packagePath string) {
	f, err := os.Open(projectPath + "/go.mod")
	if err != nil {
		elog.Error("get packagePath err", elog.FieldErr(err))
		return
	}
	defer f.Close()
	var contentByte []byte
	contentByte, err = ioutil.ReadAll(f)
	if err != nil {
		return
	}
	for _, s := range strings.Split(string(contentByte), "\n") {
		packagePath = strings.TrimSpace(strings.TrimPrefix(s, "module"))
		return
	}
	return
}

// 初始化用户配置
func (p *astParser) initOption() {
	// 如果是Go语言，那么就需要判断是否有go.mod，因为需要go.mod里的数据
	if !IsExist(p.userOption.RootPath + "/go.mod") {
		panic(fmt.Errorf("请在%s目录下创建go.mod文件", p.userOption.RootPath))
		return
	}
	p.GoMod = getPackagePath(p.userOption.RootPath)
	if p.userOption.Dependences == "*" {
		p.ParseAllDependency = true
	} else {
		p.Dependences = strings.Split(p.userOption.Dependences, ",")
	}
}

// IsExist returns whether a file or directory exists.
func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}
