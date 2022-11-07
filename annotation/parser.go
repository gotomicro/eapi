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
		return p.required()
	case "@nullable":
		return p.nullable()
	case "@consume":
		return p.consumeAnnotation()
	case "@produce":
		return p.produceAnnotation()
	case "@ignore":
		return &IgnoreAnnotation{}
	default: // unresolved plugin
		return p.unresolvedAnnotation(tag)
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

func (p *Parser) lookahead() *Token {
	if !p.hasMore() {
		return nil
	}
	return p.tokens[p.position]
}

func (p *Parser) hasMore() bool {
	return len(p.tokens) > p.position
}

func (p *Parser) required() *RequiredAnnotation {
	return &RequiredAnnotation{}
}

func (p *Parser) nullable() *NullableAnnotation {
	return &NullableAnnotation{}
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

func (p *Parser) unresolvedAnnotation(tag *Token) Annotation {
	return &UnresolvedAnnotation{
		Tag:    tag.Image,
		Tokens: p.tokens,
	}
}
