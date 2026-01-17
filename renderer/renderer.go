package renderer

import (
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/myuon/penny/css"
	"github.com/myuon/penny/dom"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type Renderer struct {
	Width  int
	Height int
}

func New(width, height int) *Renderer {
	return &Renderer{
		Width:  width,
		Height: height,
	}
}

type StyledNode struct {
	Node  *dom.Node
	Style css.Style
	Text  string
}

func (r *Renderer) Render(d *dom.DOM, stylesheet *css.Stylesheet, outputPath string) error {
	img := image.NewRGBA(image.Rect(0, 0, r.Width, r.Height))

	// Fill background with white
	for y := 0; y < r.Height; y++ {
		for x := 0; x < r.Width; x++ {
			img.Set(x, y, color.White)
		}
	}

	// Compute styles and extract styled nodes from body
	styledNodes := r.computeStyles(d, stylesheet)

	// Render styled nodes
	r.renderNodes(img, styledNodes)

	// Save to file
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}

func (r *Renderer) computeStyles(d *dom.DOM, stylesheet *css.Stylesheet) []StyledNode {
	var styledNodes []StyledNode
	var inBody bool

	var walk func(nodeID dom.NodeID, parentStyle css.Style)
	walk = func(nodeID dom.NodeID, parentStyle css.Style) {
		node := d.GetNode(nodeID)
		if node == nil {
			return
		}

		// Compute style for this node
		style := css.DefaultStyle()
		// Inherit some properties from parent
		style.Color = parentStyle.Color
		style.FontSize = parentStyle.FontSize

		if node.Type == dom.NodeTypeElement {
			// Apply matching rules
			if stylesheet != nil {
				for _, rule := range stylesheet.Rules {
					if matchesSelector(node, rule.Selectors) {
						for _, decl := range rule.Declarations {
							css.ApplyDeclaration(&style, decl)
						}
					}
				}
			}
		}

		if node.Type == dom.NodeTypeElement && node.Tag == "body" {
			inBody = true
			// Apply body background to whole page if set
		}

		if inBody && node.Type == dom.NodeTypeText {
			styledNodes = append(styledNodes, StyledNode{
				Node:  node,
				Style: parentStyle, // Use parent's style for text
				Text:  node.Text,
			})
		}

		for _, childID := range node.Children {
			walk(childID, style)
		}

		if node.Type == dom.NodeTypeElement && node.Tag == "body" {
			inBody = false
		}
	}

	walk(d.Root, css.DefaultStyle())
	return styledNodes
}

func matchesSelector(node *dom.Node, selectors []css.Selector) bool {
	for _, sel := range selectors {
		switch sel.Type {
		case css.SelectorTag:
			if node.Tag == sel.Value {
				return true
			}
		case css.SelectorClass:
			if class, ok := node.Attr["class"]; ok {
				if class == sel.Value {
					return true
				}
			}
		case css.SelectorID:
			if id, ok := node.Attr["id"]; ok {
				if id == sel.Value {
					return true
				}
			}
		}
	}
	return false
}

func (r *Renderer) renderNodes(img *image.RGBA, nodes []StyledNode) {
	face := basicfont.Face7x13

	drawer := &font.Drawer{
		Dst:  img,
		Face: face,
	}

	x := 10
	y := 20

	for _, sn := range nodes {
		// Set color from style
		col := color.RGBA{sn.Style.Color.R, sn.Style.Color.G, sn.Style.Color.B, sn.Style.Color.A}
		drawer.Src = image.NewUniform(col)

		// Calculate line height from font size
		lineHeight := int(sn.Style.FontSize * 1.5)
		if lineHeight < 15 {
			lineHeight = 15
		}

		drawer.Dot = fixed.Point26_6{
			X: fixed.I(x),
			Y: fixed.I(y),
		}
		drawer.DrawString(sn.Text)
		y += lineHeight
	}
}
