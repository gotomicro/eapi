package gin

type Config struct {
	RouterNames []string       `yaml:"routerNames"`
	Request     []RequestRule  `yaml:"request"`
	Response    []ResponseRule `yaml:"response"`
}

type ResponseRule struct {
	Node   string          `yaml:"node"`
	Type   string          `yaml:"type"` // type name
	Method string          `yaml:"method"`
	Return *ResponseReturn `yaml:"return"`
}

type ResponseReturn struct {
	Status      string      `yaml:"status"`
	Data        *DataSchema `yaml:"data"`
	ContentType string      `yaml:"contentType"`
}

type RequestRule struct {
	Type   string         `yaml:"type"` // type name
	Method string         `yaml:"method"`
	Return *RequestReturn `yaml:"return"`
}

type RequestReturn struct {
	Data        *DataSchema `yaml:"data"`
	ContentType string      `yaml:"contentType"`
}

type DataType string

const (
	DataTypeString  DataType = "string"
	DataTypeNumber  DataType = "number"
	DataTypeInteger DataType = "integer"
	DataTypeBoolean DataType = "boolean"
	DataTypeArray   DataType = "array"
	DataTypeFile    DataType = "file"
	DataTypeObject  DataType = "object"
)

type DataSchema struct {
	Type   DataType `json:"type"`
	Format string   `json:"format"`

	Properties DataProperties `json:"properties"`
	// set when Type='array'
	Item     *DataSchema `json:"item"`
	Optional bool        `json:"optional"`
}

type DataProperties map[string]*DataSchema
