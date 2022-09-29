package parser

// user option
type UserOption struct {
	GoMod              string
	RouterMod          string
	Mode               string `json:"mode"` // mode: tmpl 模板，json json数据
	ContextDebug       bool   `json:"contextDebug"`
	ScaffoldDSLContent string `json:"scaffoldDslContent"`
	RootFile           string `json:"rootFile"`
}
