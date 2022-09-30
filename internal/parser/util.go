package parser

import (
	"fmt"
	"strconv"
	"strings"
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
