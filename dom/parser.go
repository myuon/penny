package dom

import (
	"io"
	"strings"
)

// Parser builds a DOM tree from tokens
type Parser struct {
	lexer  *Lexer
	dom    *DOM
	stack  []NodeID // stack of open elements
}

func Parse(r io.Reader) (*DOM, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return ParseString(string(data))
}

func ParseString(s string) (*DOM, error) {
	parser := &Parser{
		lexer: NewLexer(s),
		dom:   NewDOM(),
		stack: []NodeID{},
	}

	parser.parse()

	return parser.dom, nil
}

func (p *Parser) parse() {
	for {
		tok := p.lexer.NextToken()
		if tok.Type == TokenEOF {
			break
		}

		switch tok.Type {
		case TokenDoctype:
			// Skip doctype for now
		case TokenComment:
			// Skip comments for now
		case TokenStartTag:
			p.handleStartTag(tok)
		case TokenEndTag:
			p.handleEndTag(tok)
		case TokenSelfClosingTag:
			p.handleSelfClosingTag(tok)
		case TokenText:
			p.handleText(tok)
		}
	}
}

func (p *Parser) currentParent() NodeID {
	if len(p.stack) == 0 {
		return InvalidNodeID
	}
	return p.stack[len(p.stack)-1]
}

func (p *Parser) handleStartTag(tok Token) {
	nodeID := p.dom.CreateElement(tok.Data)
	for _, attr := range tok.Attributes {
		p.dom.SetAttribute(nodeID, attr.Key, attr.Value)
	}

	parent := p.currentParent()
	if parent != InvalidNodeID {
		p.dom.AppendChild(parent, nodeID)
	}

	// Set root if not set
	if p.dom.Root == InvalidNodeID {
		p.dom.Root = nodeID
	}

	// Push to stack (for non-void elements)
	if !isVoidElement(tok.Data) {
		p.stack = append(p.stack, nodeID)
	}
}

func (p *Parser) handleEndTag(tok Token) {
	// Pop from stack, looking for matching tag
	for i := len(p.stack) - 1; i >= 0; i-- {
		node := p.dom.GetNode(p.stack[i])
		if node != nil && node.Tag == tok.Data {
			p.stack = p.stack[:i]
			return
		}
	}
}

func (p *Parser) handleSelfClosingTag(tok Token) {
	nodeID := p.dom.CreateElement(tok.Data)
	for _, attr := range tok.Attributes {
		p.dom.SetAttribute(nodeID, attr.Key, attr.Value)
	}

	parent := p.currentParent()
	if parent != InvalidNodeID {
		p.dom.AppendChild(parent, nodeID)
	}

	// Set root if not set
	if p.dom.Root == InvalidNodeID {
		p.dom.Root = nodeID
	}
	// Don't push to stack - self-closing
}

func (p *Parser) handleText(tok Token) {
	text := strings.TrimSpace(tok.Data)
	if text == "" {
		return // Skip whitespace-only text nodes
	}

	nodeID := p.dom.CreateText(text)

	parent := p.currentParent()
	if parent != InvalidNodeID {
		p.dom.AppendChild(parent, nodeID)
	}
}

// isVoidElement returns true for HTML void elements that don't have closing tags
func isVoidElement(tag string) bool {
	switch tag {
	case "area", "base", "br", "col", "embed", "hr", "img", "input",
		"link", "meta", "param", "source", "track", "wbr":
		return true
	}
	return false
}
