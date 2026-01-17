package dom

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenDoctype
	TokenStartTag
	TokenEndTag
	TokenSelfClosingTag
	TokenText
	TokenComment
)

func (t TokenType) String() string {
	switch t {
	case TokenEOF:
		return "EOF"
	case TokenDoctype:
		return "Doctype"
	case TokenStartTag:
		return "StartTag"
	case TokenEndTag:
		return "EndTag"
	case TokenSelfClosingTag:
		return "SelfClosingTag"
	case TokenText:
		return "Text"
	case TokenComment:
		return "Comment"
	default:
		return "Unknown"
	}
}

type Attribute struct {
	Key   string
	Value string
}

type Token struct {
	Type       TokenType
	Data       string       // tag name or text content
	Attributes []Attribute  // for start tags
}

func (t Token) String() string {
	switch t.Type {
	case TokenStartTag, TokenSelfClosingTag:
		return fmt.Sprintf("%s<%s %v>", t.Type, t.Data, t.Attributes)
	case TokenEndTag:
		return fmt.Sprintf("%s</%s>", t.Type, t.Data)
	default:
		return fmt.Sprintf("%s(%q)", t.Type, t.Data)
	}
}

type Lexer struct {
	input string
	pos   int
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input: input,
		pos:   0,
	}
}

func (l *Lexer) peek() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) peekN(n int) string {
	end := l.pos + n
	if end > len(l.input) {
		end = len(l.input)
	}
	return l.input[l.pos:end]
}

func (l *Lexer) advance() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	ch := l.input[l.pos]
	l.pos++
	return ch
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && unicode.IsSpace(rune(l.input[l.pos])) {
		l.pos++
	}
}

func (l *Lexer) NextToken() Token {
	if l.pos >= len(l.input) {
		return Token{Type: TokenEOF}
	}

	if l.peek() == '<' {
		return l.tag()
	}

	return l.text()
}

func (l *Lexer) text() Token {
	start := l.pos
	for l.pos < len(l.input) && l.peek() != '<' {
		l.pos++
	}
	text := l.input[start:l.pos]
	return Token{Type: TokenText, Data: text}
}

func (l *Lexer) tag() Token {
	l.advance() // consume '<'

	// Comment: <!-- ... -->
	if l.peekN(3) == "!--" {
		l.pos += 3 // consume "!--"
		return l.comment()
	}

	// Doctype: <!DOCTYPE ...>
	if l.peekN(8) == "!DOCTYPE" || l.peekN(8) == "!doctype" {
		l.pos += 8 // consume "!DOCTYPE"
		return l.doctype()
	}

	// End tag: </...>
	if l.peek() == '/' {
		l.advance() // consume '/'
		return l.endTag()
	}

	// Start tag or self-closing tag
	return l.startTag()
}

func (l *Lexer) comment() Token {
	start := l.pos
	for l.pos < len(l.input) {
		if l.peekN(3) == "-->" {
			content := l.input[start:l.pos]
			l.pos += 3 // consume "-->"
			return Token{Type: TokenComment, Data: content}
		}
		l.pos++
	}
	// Unclosed comment
	return Token{Type: TokenComment, Data: l.input[start:]}
}

func (l *Lexer) doctype() Token {
	l.skipWhitespace()
	start := l.pos
	for l.pos < len(l.input) && l.peek() != '>' {
		l.pos++
	}
	content := strings.TrimSpace(l.input[start:l.pos])
	if l.peek() == '>' {
		l.advance() // consume '>'
	}
	return Token{Type: TokenDoctype, Data: content}
}

func (l *Lexer) endTag() Token {
	l.skipWhitespace()
	tagName := l.tagName()
	l.skipWhitespace()
	if l.peek() == '>' {
		l.advance() // consume '>'
	}
	return Token{Type: TokenEndTag, Data: tagName}
}

func (l *Lexer) startTag() Token {
	l.skipWhitespace()
	tagName := l.tagName()
	attrs := l.attributes()

	l.skipWhitespace()

	// Check for self-closing
	if l.peek() == '/' {
		l.advance() // consume '/'
		l.skipWhitespace()
		if l.peek() == '>' {
			l.advance() // consume '>'
		}
		return Token{Type: TokenSelfClosingTag, Data: tagName, Attributes: attrs}
	}

	if l.peek() == '>' {
		l.advance() // consume '>'
	}

	return Token{Type: TokenStartTag, Data: tagName, Attributes: attrs}
}

func (l *Lexer) tagName() string {
	start := l.pos
	for l.pos < len(l.input) {
		ch := l.peek()
		if unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch)) || ch == '-' || ch == '_' {
			l.pos++
		} else {
			break
		}
	}
	return strings.ToLower(l.input[start:l.pos])
}

func (l *Lexer) attributes() []Attribute {
	var attrs []Attribute

	for {
		l.skipWhitespace()
		if l.pos >= len(l.input) || l.peek() == '>' || l.peek() == '/' {
			break
		}

		attr := l.attribute()
		if attr.Key != "" {
			attrs = append(attrs, attr)
		}
	}

	return attrs
}

func (l *Lexer) attribute() Attribute {
	// Read attribute name
	start := l.pos
	for l.pos < len(l.input) {
		ch := l.peek()
		if unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch)) || ch == '-' || ch == '_' || ch == ':' {
			l.pos++
		} else {
			break
		}
	}
	name := strings.ToLower(l.input[start:l.pos])

	if name == "" {
		return Attribute{}
	}

	l.skipWhitespace()

	// Check for '='
	if l.peek() != '=' {
		// Attribute without value
		return Attribute{Key: name, Value: ""}
	}
	l.advance() // consume '='

	l.skipWhitespace()

	// Read attribute value
	value := l.attributeValue()

	return Attribute{Key: name, Value: value}
}

func (l *Lexer) attributeValue() string {
	quote := l.peek()
	if quote == '"' || quote == '\'' {
		l.advance() // consume opening quote
		start := l.pos
		for l.pos < len(l.input) && l.peek() != quote {
			l.pos++
		}
		value := l.input[start:l.pos]
		if l.peek() == quote {
			l.advance() // consume closing quote
		}
		return value
	}

	// Unquoted value
	start := l.pos
	for l.pos < len(l.input) {
		ch := l.peek()
		if unicode.IsSpace(rune(ch)) || ch == '>' || ch == '/' {
			break
		}
		l.pos++
	}
	return l.input[start:l.pos]
}

// Tokenize returns all tokens from the input
func (l *Lexer) Tokenize() []Token {
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TokenEOF {
			break
		}
	}
	return tokens
}
