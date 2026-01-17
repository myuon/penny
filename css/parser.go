package css

import (
	"strconv"
	"strings"
)

type SelectorType int

const (
	SelectorTag SelectorType = iota
	SelectorClass
	SelectorID
)

type Selector struct {
	Type  SelectorType
	Value string
}

type Declaration struct {
	Property string
	Value    string
	Values   []Token // parsed tokens for complex values
}

type Rule struct {
	Selectors    []Selector
	Declarations []Declaration
}

type Stylesheet struct {
	Rules []Rule
}

type Parser struct {
	lexer *Lexer
	cur   Token
}

func Parse(input string) (*Stylesheet, error) {
	parser := &Parser{
		lexer: NewLexer(input),
	}
	parser.advance()
	return parser.parse(), nil
}

func (p *Parser) advance() {
	p.cur = p.lexer.NextToken()
}

func (p *Parser) parse() *Stylesheet {
	var rules []Rule
	for p.cur.Type != TokenEOF {
		rule := p.rule()
		if len(rule.Selectors) > 0 {
			rules = append(rules, rule)
		}
	}
	return &Stylesheet{Rules: rules}
}

func (p *Parser) rule() Rule {
	selectors := p.selectors()

	if p.cur.Type != TokenLBrace {
		// Skip until we find a brace or EOF
		for p.cur.Type != TokenLBrace && p.cur.Type != TokenEOF {
			p.advance()
		}
	}

	if p.cur.Type == TokenLBrace {
		p.advance() // consume '{'
	}

	declarations := p.declarations()

	if p.cur.Type == TokenRBrace {
		p.advance() // consume '}'
	}

	return Rule{
		Selectors:    selectors,
		Declarations: declarations,
	}
}

func (p *Parser) selectors() []Selector {
	var selectors []Selector

	for {
		sel := p.selector()
		if sel.Value != "" {
			selectors = append(selectors, sel)
		}

		if p.cur.Type == TokenComma {
			p.advance() // consume ','
			continue
		}
		break
	}

	return selectors
}

func (p *Parser) selector() Selector {
	switch p.cur.Type {
	case TokenIdent:
		value := p.cur.Value
		p.advance()
		return Selector{Type: SelectorTag, Value: value}
	case TokenDot:
		p.advance() // consume '.'
		if p.cur.Type == TokenIdent {
			value := p.cur.Value
			p.advance()
			return Selector{Type: SelectorClass, Value: value}
		}
	case TokenHash:
		value := p.cur.Value
		p.advance()
		return Selector{Type: SelectorID, Value: value}
	}
	return Selector{}
}

func (p *Parser) declarations() []Declaration {
	var decls []Declaration

	for p.cur.Type != TokenRBrace && p.cur.Type != TokenEOF {
		decl := p.declaration()
		if decl.Property != "" {
			decls = append(decls, decl)
		}
	}

	return decls
}

func (p *Parser) declaration() Declaration {
	if p.cur.Type != TokenIdent {
		p.advance()
		return Declaration{}
	}

	property := p.cur.Value
	p.advance()

	if p.cur.Type != TokenColon {
		return Declaration{}
	}
	p.advance() // consume ':'

	// Collect value tokens until semicolon or closing brace
	var values []Token
	var valueStr strings.Builder

	for p.cur.Type != TokenSemicolon && p.cur.Type != TokenRBrace && p.cur.Type != TokenEOF {
		values = append(values, p.cur)
		if valueStr.Len() > 0 {
			valueStr.WriteString(" ")
		}
		valueStr.WriteString(p.cur.Value)
		if p.cur.Unit != "" {
			valueStr.WriteString(p.cur.Unit)
		}
		p.advance()
	}

	if p.cur.Type == TokenSemicolon {
		p.advance() // consume ';'
	}

	return Declaration{
		Property: property,
		Value:    valueStr.String(),
		Values:   values,
	}
}

// ApplyDeclaration applies a CSS declaration to a Style
func ApplyDeclaration(style *Style, decl Declaration) {
	switch decl.Property {
	case "display":
		switch decl.Value {
		case "block":
			style.Display = DisplayBlock
		case "inline":
			style.Display = DisplayInline
		case "none":
			style.Display = DisplayNone
		case "flex":
			style.Display = DisplayFlex
		}

	case "width":
		if v := parseLength(decl.Values); v != nil {
			style.Width = v
		}
	case "height":
		if v := parseLength(decl.Values); v != nil {
			style.Height = v
		}

	case "margin":
		style.Margin = parseEdges(decl.Values)
	case "margin-top":
		if v := parseLength(decl.Values); v != nil {
			style.Margin.Top = *v
		}
	case "margin-right":
		if v := parseLength(decl.Values); v != nil {
			style.Margin.Right = *v
		}
	case "margin-bottom":
		if v := parseLength(decl.Values); v != nil {
			style.Margin.Bottom = *v
		}
	case "margin-left":
		if v := parseLength(decl.Values); v != nil {
			style.Margin.Left = *v
		}

	case "padding":
		style.Padding = parseEdges(decl.Values)
	case "padding-top":
		if v := parseLength(decl.Values); v != nil {
			style.Padding.Top = *v
		}
	case "padding-right":
		if v := parseLength(decl.Values); v != nil {
			style.Padding.Right = *v
		}
	case "padding-bottom":
		if v := parseLength(decl.Values); v != nil {
			style.Padding.Bottom = *v
		}
	case "padding-left":
		if v := parseLength(decl.Values); v != nil {
			style.Padding.Left = *v
		}

	case "font-size":
		if v := parseLength(decl.Values); v != nil {
			style.FontSize = *v
		}

	case "color":
		if c := parseColor(decl); c != nil {
			style.Color = *c
		}

	case "background", "background-color":
		if c := parseColor(decl); c != nil {
			style.Background = *c
		}

	case "border-width":
		style.Border = parseEdges(decl.Values)

	case "border-color":
		if c := parseColor(decl); c != nil {
			style.BorderColor = *c
		}

	case "flex-grow":
		if len(decl.Values) > 0 && decl.Values[0].Type == TokenNumber {
			if v, err := strconv.ParseFloat(decl.Values[0].Value, 32); err == nil {
				style.FlexGrow = float32(v)
			}
		}

	case "justify-content":
		switch decl.Value {
		case "flex-start":
			style.JustifyContent = JustifyFlexStart
		case "flex-end":
			style.JustifyContent = JustifyFlexEnd
		case "center":
			style.JustifyContent = JustifyCenter
		case "space-between":
			style.JustifyContent = JustifySpaceBetween
		case "space-around":
			style.JustifyContent = JustifySpaceAround
		}

	case "align-items":
		switch decl.Value {
		case "flex-start":
			style.AlignItems = AlignFlexStart
		case "flex-end":
			style.AlignItems = AlignFlexEnd
		case "center":
			style.AlignItems = AlignCenter
		case "stretch":
			style.AlignItems = AlignStretch
		}
	}
}

func parseLength(values []Token) *float32 {
	if len(values) == 0 {
		return nil
	}

	tok := values[0]
	var v float64
	var err error

	switch tok.Type {
	case TokenNumber:
		v, err = strconv.ParseFloat(tok.Value, 32)
	case TokenDimension:
		v, err = strconv.ParseFloat(tok.Value, 32)
		// For now, treat all units as pixels
		// TODO: handle em, rem, etc.
	case TokenPercentage:
		// TODO: handle percentage properly
		v, err = strconv.ParseFloat(tok.Value, 32)
	default:
		return nil
	}

	if err != nil {
		return nil
	}

	f := float32(v)
	return &f
}

func parseEdges(values []Token) Edges {
	var lengths []float32
	for _, tok := range values {
		if tok.Type == TokenNumber || tok.Type == TokenDimension {
			if v, err := strconv.ParseFloat(tok.Value, 32); err == nil {
				lengths = append(lengths, float32(v))
			}
		}
	}

	switch len(lengths) {
	case 1:
		return Edges{lengths[0], lengths[0], lengths[0], lengths[0]}
	case 2:
		return Edges{lengths[0], lengths[1], lengths[0], lengths[1]}
	case 3:
		return Edges{lengths[0], lengths[1], lengths[2], lengths[1]}
	case 4:
		return Edges{lengths[0], lengths[1], lengths[2], lengths[3]}
	default:
		return Edges{}
	}
}

func parseColor(decl Declaration) *Color {
	// Handle named colors
	switch decl.Value {
	case "black":
		return &Color{0, 0, 0, 255}
	case "white":
		return &Color{255, 255, 255, 255}
	case "red":
		return &Color{255, 0, 0, 255}
	case "green":
		return &Color{0, 128, 0, 255}
	case "blue":
		return &Color{0, 0, 255, 255}
	case "yellow":
		return &Color{255, 255, 0, 255}
	case "gray", "grey":
		return &Color{128, 128, 128, 255}
	case "transparent":
		return &Color{0, 0, 0, 0}
	}

	// Handle #hex
	if len(decl.Values) > 0 && decl.Values[0].Type == TokenHash {
		hex := decl.Values[0].Value
		return parseHexColor(hex)
	}

	// Handle rgb() / rgba()
	if len(decl.Values) > 0 && decl.Values[0].Type == TokenFunction {
		fn := decl.Values[0].Value
		if fn == "rgb" || fn == "rgba" {
			return parseRGBFunction(decl.Values[1:])
		}
	}

	return nil
}

func parseHexColor(hex string) *Color {
	hex = strings.TrimPrefix(hex, "#")

	var r, g, b, a uint8 = 0, 0, 0, 255

	switch len(hex) {
	case 3: // #RGB
		r = parseHexByte(hex[0:1] + hex[0:1])
		g = parseHexByte(hex[1:2] + hex[1:2])
		b = parseHexByte(hex[2:3] + hex[2:3])
	case 6: // #RRGGBB
		r = parseHexByte(hex[0:2])
		g = parseHexByte(hex[2:4])
		b = parseHexByte(hex[4:6])
	case 8: // #RRGGBBAA
		r = parseHexByte(hex[0:2])
		g = parseHexByte(hex[2:4])
		b = parseHexByte(hex[4:6])
		a = parseHexByte(hex[6:8])
	default:
		return nil
	}

	return &Color{r, g, b, a}
}

func parseHexByte(s string) uint8 {
	v, _ := strconv.ParseUint(s, 16, 8)
	return uint8(v)
}

func parseRGBFunction(values []Token) *Color {
	var nums []uint8
	for _, tok := range values {
		if tok.Type == TokenNumber {
			if v, err := strconv.ParseUint(tok.Value, 10, 8); err == nil {
				nums = append(nums, uint8(v))
			}
		}
	}

	if len(nums) >= 3 {
		a := uint8(255)
		if len(nums) >= 4 {
			a = nums[3]
		}
		return &Color{nums[0], nums[1], nums[2], a}
	}

	return nil
}

func (s *Stylesheet) Dump() string {
	var result string
	for _, rule := range s.Rules {
		// Selectors
		for i, sel := range rule.Selectors {
			if i > 0 {
				result += ", "
			}
			switch sel.Type {
			case SelectorTag:
				result += sel.Value
			case SelectorClass:
				result += "." + sel.Value
			case SelectorID:
				result += "#" + sel.Value
			}
		}
		result += " {\n"

		// Declarations
		for _, decl := range rule.Declarations {
			result += "  " + decl.Property + ": " + decl.Value + ";\n"
		}
		result += "}\n"
	}
	return result
}
