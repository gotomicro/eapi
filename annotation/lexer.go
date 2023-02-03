package annotation

import (
	"fmt"
	"regexp"
)

type TokenType int

const (
	tokenTag TokenType = iota + 1
	tokenString
	tokenNumber
	tokenBool
	tokenWhiteSpace
	tokenIdentifier
)

var patterns = []*pattern{
	newPattern(tokenTag, "^@[a-zA-Z_]+\\w*"),
	newPattern(tokenString, "^\"(\\\\.|[^\"])*\""),
	newPattern(tokenNumber, "^[+-]?([0-9]*[.])?[0-9]+"),
	newPattern(tokenBool, "/^(true|false)/i"),
	newPattern(tokenWhiteSpace, "^\\s+"),
	newPattern(tokenIdentifier, "^[^\\s]+"),
}

var tokenNameMap = map[TokenType]string{
	tokenTag:        "tag",
	tokenString:     "string",
	tokenNumber:     "number",
	tokenBool:       "bool",
	tokenWhiteSpace: "whitespace",
	tokenIdentifier: "identifier",
}

type pattern struct {
	tokenType TokenType
	pattern   *regexp.Regexp
}

func newPattern(tokenType TokenType, exp string) *pattern {
	return &pattern{tokenType: tokenType, pattern: regexp.MustCompile(exp)}
}

type Token struct {
	Type  TokenType
	Image string
}

type Lexer struct {
	code string
}

func NewLexer(code string) *Lexer {
	return &Lexer{code: code}
}

func (l *Lexer) Lex() (tokens []*Token, err error) {
	cursor := 0
	for cursor < len(l.code) {
		var matched string
		for _, p := range patterns {
			matched = p.pattern.FindString(l.code[cursor:])
			if matched != "" {
				tokens = append(tokens, &Token{
					Type:  p.tokenType,
					Image: matched,
				})
				break
			}
		}
		if matched == "" {
			err = fmt.Errorf("unexpected token after %s", l.code[:cursor])
			return
		}

		cursor += len(matched)
	}
	return
}
