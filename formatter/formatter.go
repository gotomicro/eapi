package formatter

import "strings"

type Options struct {
	IndentWidth int
}

func Format(doc Doc, options *Options) string {
	if options == nil {
		options = &Options{IndentWidth: 4}
	}
	return NewFormatter(options).Format(doc)
}

type Formatter struct {
	options     *Options
	indent      int
	atLineStart bool
}

func NewFormatter(options *Options) *Formatter {
	return &Formatter{options: options, atLineStart: true}
}

func (f *Formatter) Format(doc Doc) string {
	var code string

	switch doc := doc.(type) {
	case *DocContent:
		if f.atLineStart {
			code += f.makeIndent()
			f.atLineStart = false
		}
		code += string(*doc)

	case *DocGroup:
		for _, doc := range doc.Docs {
			code += f.Format(doc)
		}

	case *DocNode:
		code = f.docNode(doc)
	}

	return code
}

func (f *Formatter) docNode(doc *DocNode) string {
	switch doc.Type {
	case DocNodeIndent:
		f.indent += 1
		defer func() { f.indent -= 1 }()
		return f.Format(doc.Child)

	case DocNodeLineBreak:
		res := "\n"
		f.atLineStart = true
		return res
	}

	return ""
}

func (f *Formatter) makeIndent() string {
	return strings.Repeat(" ", f.indent*f.options.IndentWidth)
}
