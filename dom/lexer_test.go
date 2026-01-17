package dom

import (
	"testing"
)

func TestLexer(t *testing.T) {
	input := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Test</title>
</head>
<body>
  <div class="container">
    <h1>Hello, World!</h1>
    <p>This is a test.</p>
    <br/>
  </div>
</body>
</html>`

	lexer := NewLexer(input)
	tokens := lexer.Tokenize()

	for _, tok := range tokens {
		t.Logf("%v", tok)
	}

	// Basic checks
	if tokens[0].Type != TokenDoctype {
		t.Errorf("expected Doctype, got %v", tokens[0].Type)
	}
	if tokens[0].Data != "html" {
		t.Errorf("expected 'html', got %q", tokens[0].Data)
	}
}

func TestLexerStartTag(t *testing.T) {
	input := `<div class="foo" id="bar">`
	lexer := NewLexer(input)
	tok := lexer.NextToken()

	if tok.Type != TokenStartTag {
		t.Errorf("expected StartTag, got %v", tok.Type)
	}
	if tok.Data != "div" {
		t.Errorf("expected 'div', got %q", tok.Data)
	}
	if len(tok.Attributes) != 2 {
		t.Errorf("expected 2 attributes, got %d", len(tok.Attributes))
	}
	if tok.Attributes[0].Key != "class" || tok.Attributes[0].Value != "foo" {
		t.Errorf("unexpected attribute: %v", tok.Attributes[0])
	}
}

func TestLexerSelfClosing(t *testing.T) {
	input := `<br/>`
	lexer := NewLexer(input)
	tok := lexer.NextToken()

	if tok.Type != TokenSelfClosingTag {
		t.Errorf("expected SelfClosingTag, got %v", tok.Type)
	}
	if tok.Data != "br" {
		t.Errorf("expected 'br', got %q", tok.Data)
	}
}

func TestLexerComment(t *testing.T) {
	input := `<!-- this is a comment -->`
	lexer := NewLexer(input)
	tok := lexer.NextToken()

	if tok.Type != TokenComment {
		t.Errorf("expected Comment, got %v", tok.Type)
	}
	if tok.Data != " this is a comment " {
		t.Errorf("unexpected comment content: %q", tok.Data)
	}
}
