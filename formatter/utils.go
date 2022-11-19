package formatter

func Indent(doc Doc) *DocNode {
	return &DocNode{
		Type:  DocNodeIndent,
		Child: doc,
	}
}

func LineBreak() *DocNode {
	return &DocNode{
		Type: DocNodeLineBreak,
	}
}

func Join(sep Doc, docs ...Doc) *DocGroup {
	res := &DocGroup{
		Docs: make([]Doc, 0),
	}
	for i, item := range docs {
		if i > 0 {
			res.Docs = append(res.Docs, sep)
		}
		res.Docs = append(res.Docs, item)
	}
	return res
}

func Group(docs ...Doc) *DocGroup {
	return &DocGroup{Docs: docs}
}

func Content(code string) *DocContent {
	res := DocContent(code)
	return &res
}
