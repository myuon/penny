package css

import (
	"unicode"
)

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdent      // property name, tag name, class name
	TokenHash       // #id
	TokenDot        // .
	TokenColon      // :
	TokenSemicolon  // ;
	TokenComma      // ,
	TokenLBrace     // {
	TokenRBrace     // }
	TokenNumber     // 123, 12.5
	TokenDimension  // 10px, 2em
	TokenPercentage // 50%
	TokenString     // "..." or '...'
	TokenFunction   // rgb(
	TokenRParen     // )
)

func (t TokenType) String() string {
	switch t {
	case TokenEOF:
		return "EOF"
	case TokenIdent:
		return "Ident"
	case TokenHash:
		return "Hash"
	case TokenDot:
		return "Dot"
	case TokenColon:
		return "Colon"
	case TokenSemicolon:
		return "Semicolon"
	case TokenComma:
		return "Comma"
	case TokenLBrace:
		return "LBrace"
	case TokenRBrace:
		return "RBrace"
	case TokenNumber:
		return "Number"
	case TokenDimension:
		return "Dimension"
	case TokenPercentage:
		return "Percentage"
	case TokenString:
		return "String"
	case TokenFunction:
		return "Function"
	case TokenRParen:
		return "RParen"
	default:
		return "Unknown"
	}
}

type Token struct {
	Type  TokenType
	Value string
	Unit  string // for Dimension
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

func (l *Lexer) advance() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	ch := l.input[l.pos]
	l.pos++
	return ch
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		ch := l.peek()
		if unicode.IsSpace(rune(ch)) {
			l.pos++
		} else if ch == '/' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '*' {
			// Skip /* ... */ comments
			l.pos += 2
			for l.pos+1 < len(l.input) {
				if l.input[l.pos] == '*' && l.input[l.pos+1] == '/' {
					l.pos += 2
					break
				}
				l.pos++
			}
		} else {
			break
		}
	}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return Token{Type: TokenEOF}
	}

	ch := l.peek()

	switch ch {
	case '{':
		l.advance()
		return Token{Type: TokenLBrace, Value: "{"}
	case '}':
		l.advance()
		return Token{Type: TokenRBrace, Value: "}"}
	case ':':
		l.advance()
		return Token{Type: TokenColon, Value: ":"}
	case ';':
		l.advance()
		return Token{Type: TokenSemicolon, Value: ";"}
	case ',':
		l.advance()
		return Token{Type: TokenComma, Value: ","}
	case '.':
		l.advance()
		return Token{Type: TokenDot, Value: "."}
	case ')':
		l.advance()
		return Token{Type: TokenRParen, Value: ")"}
	case '#':
		return l.hash()
	case '"', '\'':
		return l.str()
	}

	if ch == '-' || unicode.IsDigit(rune(ch)) {
		return l.number()
	}

	if isIdentStart(ch) {
		return l.ident()
	}

	// Skip unknown character
	l.advance()
	return l.NextToken()
}

func (l *Lexer) hash() Token {
	l.advance() // consume '#'
	start := l.pos
	for l.pos < len(l.input) && isIdentChar(l.peek()) {
		l.pos++
	}
	return Token{Type: TokenHash, Value: l.input[start:l.pos]}
}

func (l *Lexer) str() Token {
	quote := l.advance()
	start := l.pos
	for l.pos < len(l.input) && l.peek() != quote {
		l.pos++
	}
	value := l.input[start:l.pos]
	if l.peek() == quote {
		l.advance()
	}
	return Token{Type: TokenString, Value: value}
}

func (l *Lexer) number() Token {
	start := l.pos

	// Handle negative
	if l.peek() == '-' {
		l.advance()
	}

	// Integer part
	for l.pos < len(l.input) && unicode.IsDigit(rune(l.peek())) {
		l.pos++
	}

	// Decimal part
	if l.peek() == '.' && l.pos+1 < len(l.input) && unicode.IsDigit(rune(l.input[l.pos+1])) {
		l.advance() // consume '.'
		for l.pos < len(l.input) && unicode.IsDigit(rune(l.peek())) {
			l.pos++
		}
	}

	value := l.input[start:l.pos]

	// Check for percentage
	if l.peek() == '%' {
		l.advance()
		return Token{Type: TokenPercentage, Value: value}
	}

	// Check for unit (dimension)
	if isIdentStart(l.peek()) {
		unitStart := l.pos
		for l.pos < len(l.input) && isIdentChar(l.peek()) {
			l.pos++
		}
		unit := l.input[unitStart:l.pos]
		return Token{Type: TokenDimension, Value: value, Unit: unit}
	}

	return Token{Type: TokenNumber, Value: value}
}

func (l *Lexer) ident() Token {
	start := l.pos
	for l.pos < len(l.input) && isIdentChar(l.peek()) {
		l.pos++
	}
	value := l.input[start:l.pos]

	// Check for function
	if l.peek() == '(' {
		l.advance()
		return Token{Type: TokenFunction, Value: value}
	}

	return Token{Type: TokenIdent, Value: value}
}

func isIdentStart(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_' || ch == '-'
}

func isIdentChar(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch)) || ch == '_' || ch == '-'
}

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
