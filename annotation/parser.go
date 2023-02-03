package annotation

import (
	"fmt"
	"strings"
)

type ParseError struct {
	Column  int
	Message string
}

func (e *ParseError) Error() string {
	return e.Message
}

func NewParseError(column int, message string) *ParseError {
	return &ParseError{Column: column, Message: message}
}

type Parser struct {
	text string

	tokens   []*Token
	position int
	column   int
}

func NewParser(text string) *Parser {
	return &Parser{text: text}
}

func (p *Parser) Parse() (Annotation, error) {
	var column = 0
	var text = p.text
	if strings.HasPrefix(text, "//") {
		column = 2
		text = strings.TrimPrefix(text, "//")
	}

	tokens, err := NewLexer(text).Lex()
	if err != nil {
		return nil, nil
	}
	if len(tokens) == 0 {
		return nil, nil
	}

	p.tokens = tokens
	p.position = 0
	p.column = column

	return p.parse()
}

func (p *Parser) parse() (Annotation, error) {
	tag, err := p.consume(tokenTag)
	if err != nil {
		return nil, nil
	}

	switch strings.ToLower(tag.Image) {
	case "@required":
		return newSimpleAnnotation(Required), nil
	case "@consume":
		return p.consumeAnnotation()
	case "@produce":
		return p.produceAnnotation()
	case "@ignore":
		return newSimpleAnnotation(Ignore), nil
	case "@tag", "@tags":
		return p.tags(), nil
	case "@description":
		return p.description(), nil
	case "@summary":
		return p.summary(), nil
	case "@id":
		return p.id(), nil
	case "@deprecated":
		return newSimpleAnnotation(Deprecated), nil
	case "@security":
		return p.security()
	default: // unresolved plugin
		return p.unresolved(tag), nil
	}
}

func (p *Parser) consume(typ TokenType) (*Token, error) {
	for {
		t := p.lookahead()
		if t != nil && t.Type == tokenWhiteSpace {
			p.position += 1
			p.column += len(t.Image)
		} else {
			break
		}
	}

	t := p.lookahead()
	if t == nil {
		return nil, NewParseError(p.column, fmt.Sprintf("expect %s, but got EOF", tokenNameMap[typ]))
	}
	if t.Type != typ {
		return nil, NewParseError(p.column, fmt.Sprintf("expect %s, but got '%s'", tokenNameMap[typ], t.Image))
	}

	p.position += 1
	p.column += len(t.Image)
	return t, nil
}

func (p *Parser) consumeAny() *Token {
	t := p.lookahead()
	if t == nil {
		return nil
	}

	p.position += 1
	p.column += len(t.Image)
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

func (p *Parser) consumeAnnotation() (*ConsumeAnnotation, error) {
	ident, err := p.consume(tokenIdentifier)
	if err != nil {
		return nil, err
	}
	return &ConsumeAnnotation{
		ContentType: ident.Image,
	}, nil
}

func (p *Parser) produceAnnotation() (*ProduceAnnotation, error) {
	ident, err := p.consume(tokenIdentifier)
	if err != nil {
		return nil, err
	}
	if ident == nil {
		return nil, nil
	}
	return &ProduceAnnotation{
		ContentType: ident.Image,
	}, nil
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
		ident, _ := p.consume(tokenIdentifier)
		if ident != nil {
			tag = append(tag, ident.Image)
		}
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

// @security name scope1 [...]
func (p *Parser) security() (*SecurityAnnotation, error) {
	name, err := p.consume(tokenIdentifier)
	if err != nil {
		return nil, NewParseError(p.column, "expect name after @security")
	}
	var security = SecurityAnnotation{
		Name:   name.Image,
		Params: make([]string, 0),
	}
	for p.hasMore() {
		token := p.consumeAny()
		if token.Type == tokenIdentifier {
			security.Params = append(security.Params, token.Image)
		}
	}

	return &security, nil
}
