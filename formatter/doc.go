package formatter

type Doc interface {
	docNode()
}

type DocContent string

func (d DocContent) docNode() {}

func NewDocContent(code string) *DocContent {
	var res = DocContent(code)
	return &res
}

type DocNodeType string

const (
	DocNodeIndent    DocNodeType = "string"
	DocNodeLineBreak DocNodeType = "linebreak"
)

type DocNode struct {
	Type  DocNodeType
	Child Doc
}

func (d DocNode) docNode() {}

type DocGroup struct {
	Docs []Doc
}

func (d DocGroup) docNode() {}
