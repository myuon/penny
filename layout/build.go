package layout

import (
	"github.com/myuon/penny/css"
	"github.com/myuon/penny/dom"
)

// BuildLayoutTree creates a layout tree from DOM and computed styles
// Only builds from <body> element
func BuildLayoutTree(d *dom.DOM, stylesheet *css.Stylesheet) *LayoutTree {
	tree := NewLayoutTree()

	// Find body element
	bodyID := findBody(d, d.Root)
	if bodyID == dom.InvalidNodeID {
		return tree
	}

	var build func(nodeID dom.NodeID, parentStyle css.Style) LayoutNodeID
	build = func(nodeID dom.NodeID, parentStyle css.Style) LayoutNodeID {
		node := d.GetNode(nodeID)
		if node == nil {
			return InvalidLayoutNodeID
		}

		// Compute style
		style := computeStyle(node, parentStyle, stylesheet)

		// Skip display:none
		if style.Display == css.DisplayNone {
			return InvalidLayoutNodeID
		}

		// Create layout node
		layoutID := tree.CreateNode(nodeID, style)

		// Set text for text nodes
		if node.Type == dom.NodeTypeText {
			tree.Nodes[layoutID].Text = node.Text
		}

		// Build children
		for _, childID := range node.Children {
			childLayoutID := build(childID, style)
			if childLayoutID != InvalidLayoutNodeID {
				tree.AppendChild(layoutID, childLayoutID)
			}
		}

		return layoutID
	}

	tree.Root = build(bodyID, css.DefaultStyle())
	return tree
}

func findBody(d *dom.DOM, nodeID dom.NodeID) dom.NodeID {
	node := d.GetNode(nodeID)
	if node == nil {
		return dom.InvalidNodeID
	}

	if node.Type == dom.NodeTypeElement && node.Tag == "body" {
		return nodeID
	}

	for _, childID := range node.Children {
		if found := findBody(d, childID); found != dom.InvalidNodeID {
			return found
		}
	}

	return dom.InvalidNodeID
}

func computeStyle(node *dom.Node, parentStyle css.Style, stylesheet *css.Stylesheet) css.Style {
	style := css.DefaultStyle()

	// Inherit from parent
	style.Color = parentStyle.Color
	style.FontSize = parentStyle.FontSize

	if node.Type != dom.NodeTypeElement {
		return style
	}

	// Apply matching rules
	if stylesheet == nil {
		return style
	}

	for _, rule := range stylesheet.Rules {
		if matchesSelector(node, rule.Selectors) {
			for _, decl := range rule.Declarations {
				css.ApplyDeclaration(&style, decl)
			}
		}
	}

	return style
}

func matchesSelector(node *dom.Node, selectors []css.Selector) bool {
	for _, sel := range selectors {
		switch sel.Type {
		case css.SelectorTag:
			if node.Tag == sel.Value {
				return true
			}
		case css.SelectorClass:
			if class, ok := node.Attr["class"]; ok && class == sel.Value {
				return true
			}
		case css.SelectorID:
			if id, ok := node.Attr["id"]; ok && id == sel.Value {
				return true
			}
		}
	}
	return false
}
