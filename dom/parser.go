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

// hasTagInStack returns true if the given tag exists in the stack
func (p *Parser) hasTagInStack(tag string) bool {
	for _, nodeID := range p.stack {
		node := p.dom.GetNode(nodeID)
		if node != nil && node.Tag == tag {
			return true
		}
	}
	return false
}

// ensureHtmlHead ensures that <html> and <head> elements exist in the DOM
// This is called when encountering head content without proper wrappers
func (p *Parser) ensureHtmlHead() {
	// Create <html> if not present
	if !p.hasTagInStack("html") && p.dom.Root == InvalidNodeID {
		htmlID := p.dom.CreateElement("html")
		p.dom.Root = htmlID
		p.stack = append(p.stack, htmlID)
	}

	// Create <head> if not present
	if !p.hasTagInStack("head") {
		headID := p.dom.CreateElement("head")
		parent := p.currentParent()
		if parent != InvalidNodeID {
			p.dom.AppendChild(parent, headID)
		}
		p.stack = append(p.stack, headID)
	}
}

// closeHead closes the <head> element if it's currently open
func (p *Parser) closeHead() {
	for i := len(p.stack) - 1; i >= 0; i-- {
		node := p.dom.GetNode(p.stack[i])
		if node != nil && node.Tag == "head" {
			p.stack = p.stack[:i]
			return
		}
	}
}

// ensureHtmlBody ensures that <html> and <body> elements exist in the DOM
// This is called when encountering body content without proper wrappers
func (p *Parser) ensureHtmlBody() {
	// Create <html> if not present
	if !p.hasTagInStack("html") && p.dom.Root == InvalidNodeID {
		htmlID := p.dom.CreateElement("html")
		p.dom.Root = htmlID
		p.stack = append(p.stack, htmlID)
	}

	// Close <head> if it's open (transitioning from head to body)
	if p.hasTagInStack("head") {
		p.closeHead()
	}

	// Create <body> if not present
	if !p.hasTagInStack("body") {
		bodyID := p.dom.CreateElement("body")
		parent := p.currentParent()
		if parent != InvalidNodeID {
			p.dom.AppendChild(parent, bodyID)
		}
		p.stack = append(p.stack, bodyID)
	}
}

func (p *Parser) handleStartTag(tok Token) {
	tag := tok.Data

	// Auto-insert html/head for head content elements
	if isHeadContent(tag) && !p.hasTagInStack("head") && !p.hasTagInStack("body") {
		p.ensureHtmlHead()
	}

	// Auto-insert html/body for body content elements
	if isBodyContent(tag) && !p.hasTagInStack("body") {
		p.ensureHtmlBody()
	}

	nodeID := p.dom.CreateElement(tag)
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
	if !isVoidElement(tag) {
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
	tag := tok.Data

	// Auto-insert html/head for head content elements
	if isHeadContent(tag) && !p.hasTagInStack("head") && !p.hasTagInStack("body") {
		p.ensureHtmlHead()
	}

	// Auto-insert html/body for body content elements
	if isBodyContent(tag) && !p.hasTagInStack("body") {
		p.ensureHtmlBody()
	}

	nodeID := p.dom.CreateElement(tag)
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

// isBodyContent returns true for elements that should be inside <body>
func isBodyContent(tag string) bool {
	switch tag {
	case "p", "div", "span", "h1", "h2", "h3", "h4", "h5", "h6",
		"ul", "ol", "li", "dl", "dt", "dd",
		"table", "tr", "td", "th", "thead", "tbody", "tfoot",
		"form", "fieldset", "legend", "label", "button", "select", "option", "textarea",
		"article", "section", "nav", "aside", "header", "footer", "main",
		"figure", "figcaption", "blockquote", "pre", "code",
		"a", "strong", "em", "b", "i", "u", "s", "small", "mark", "sub", "sup",
		"img", "video", "audio", "canvas", "svg", "iframe":
		return true
	}
	return false
}

// isHeadContent returns true for elements that should be inside <head>
func isHeadContent(tag string) bool {
	switch tag {
	case "title", "meta", "link", "style", "script", "base":
		return true
	}
	return false
}
