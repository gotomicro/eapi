package annotation

import "strings"

type Parser struct {
	text string

	tokens   []*Token
	position int
}

func NewParser(text string) *Parser {
	text = strings.TrimPrefix(text, "//")
	return &Parser{text: text}
}

func (p *Parser) Parse() Annotation {
	tokens, err := NewLexer(p.text).Lex()
	if err != nil {
		return nil
	}
	if len(tokens) == 0 {
		return nil
	}

	p.tokens = tokens
	p.position = 0

	return p.parse()
}

func (p *Parser) parse() Annotation {
	tag := p.consume(tokenTag)
	if tag == nil {
		return nil
	}

	switch strings.ToLower(tag.Image) {
	case "@required":
		return newSimpleAnnotation(Required)
	case "@consume":
		return p.consumeAnnotation()
	case "@produce":
		return p.produceAnnotation()
	case "@ignore":
		return newSimpleAnnotation(Ignore)
	case "@tag", "@tags":
		return p.tags()
	case "@description":
		return p.description()
	case "@summary":
		return p.summary()
	case "@id":
		return p.id()
	case "@deprecated":
		return newSimpleAnnotation(Deprecated)
	default: // unresolved plugin
		return p.unresolved(tag)
	}
}

func (p *Parser) consume(typ TokenType) *Token {
	for {
		t := p.lookahead()
		if t != nil && t.Type == tokenWhiteSpace {
			p.position += 1
		} else {
			break
		}
	}

	t := p.lookahead()
	if t == nil || t.Type != typ {
		return nil
	}

	p.position += 1
	return t
}

func (p *Parser) consumeAny() *Token {
	t := p.lookahead()
	if t == nil {
		return nil
	}

	p.position += 1
	return t
}

func (p *Parser) lookahead() *Token {
	if !p.hasMore() {
		return nil
	}
	return p.tokens[p.position]
}

func (p *Parser) hasMore() bool {
	return len(p.tokens) > p.position
}

func (p *Parser) consumeAnnotation() *ConsumeAnnotation {
	ident := p.consume(tokenIdentifier)
	if ident == nil {
		return nil
	}
	return &ConsumeAnnotation{
		ContentType: ident.Image,
	}
}

func (p *Parser) produceAnnotation() *ProduceAnnotation {
	ident := p.consume(tokenIdentifier)
	if ident == nil {
		return nil
	}
	return &ProduceAnnotation{
		ContentType: ident.Image,
	}
}

func (p *Parser) unresolved(tag *Token) Annotation {
	return &UnresolvedAnnotation{
		Tag:    tag.Image,
		Tokens: p.tokens,
	}
}

func (p *Parser) tags() Annotation {
	res := &TagAnnotation{}
	var tag []string
	for p.hasMore() {
		ident := p.consume(tokenIdentifier)
		tag = append(tag, ident.Image)
	}
	res.Tag = strings.Join(tag, " ")
	return res
}

func (p *Parser) description() Annotation {
	res := &DescriptionAnnotation{}
	for p.hasMore() {
		token := p.consumeAny()
		res.Text += token.Image
	}
	return res
}

func (p *Parser) summary() Annotation {
	res := &SummaryAnnotation{}
	for p.hasMore() {
		token := p.consumeAny()
		res.Text += token.Image
	}
	return res
}

func (p *Parser) id() Annotation {
	res := &IdAnnotation{}
	for p.hasMore() {
		token := p.consumeAny()
		res.Text += token.Image
	}
	return res
}
