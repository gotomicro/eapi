package gin

type Config struct {
	Request  []RequestRule  `yaml:"request"`
	Response []ResponseRule `yaml:"response"`
}

type ResponseRule struct {
	Node   string          `yaml:"node"`
	Type   string          `yaml:"type"` // type name
	Method string          `yaml:"method"`
	Return *ResponseReturn `yaml:"return"`
}

type ResponseReturn struct {
	Status      string `yaml:"status"`
	Data        string `yaml:"data"`
	ContentType string `yaml:"contentType"`
}

type RequestRule struct {
	Type   string         `yaml:"type"` // type name
	Method string         `yaml:"method"`
	Return *RequestReturn `yaml:"return"`
}

type RequestReturn struct {
	Data        string `yaml:"data"`
	ContentType string `yaml:"contentType"`
}
