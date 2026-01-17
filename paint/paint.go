package paint

import (
	"github.com/myuon/penny/css"
	"github.com/myuon/penny/layout"
)

// Paint generates paint operations from a layout tree
func Paint(tree *layout.LayoutTree) *PaintList {
	list := NewPaintList()

	if tree.Root == layout.InvalidLayoutNodeID {
		return list
	}

	paintNode(tree, tree.Root, list)
	return list
}

func paintNode(tree *layout.LayoutTree, nodeID layout.LayoutNodeID, list *PaintList) {
	node := tree.GetNode(nodeID)
	if node == nil {
		return
	}

	// Paint background
	if node.Style.Background.A > 0 {
		list.PushFillRect(node.Rect, node.Style.Background)
	}

	// Paint border
	if node.Style.Border.Top > 0 || node.Style.Border.Right > 0 ||
		node.Style.Border.Bottom > 0 || node.Style.Border.Left > 0 {
		paintBorder(node, list)
	}

	// Paint text
	if node.Text != "" {
		textRect := layout.Rect{
			X: node.Rect.X + node.Style.Padding.Left,
			Y: node.Rect.Y + node.Style.Padding.Top,
			W: node.Rect.W - node.Style.Padding.Left - node.Style.Padding.Right,
			H: node.Rect.H - node.Style.Padding.Top - node.Style.Padding.Bottom,
		}
		list.PushDrawText(textRect, node.Text, node.Style.Color, node.Style.FontSize)
	}

	// Paint children
	for _, childID := range node.Children {
		paintNode(tree, childID, list)
	}
}

func paintBorder(node *layout.LayoutNode, list *PaintList) {
	rect := node.Rect
	color := node.Style.BorderColor
	border := node.Style.Border

	// Top border
	if border.Top > 0 {
		list.PushFillRect(layout.Rect{
			X: rect.X,
			Y: rect.Y,
			W: rect.W,
			H: border.Top,
		}, color)
	}

	// Right border
	if border.Right > 0 {
		list.PushFillRect(layout.Rect{
			X: rect.X + rect.W - border.Right,
			Y: rect.Y,
			W: border.Right,
			H: rect.H,
		}, color)
	}

	// Bottom border
	if border.Bottom > 0 {
		list.PushFillRect(layout.Rect{
			X: rect.X,
			Y: rect.Y + rect.H - border.Bottom,
			W: rect.W,
			H: border.Bottom,
		}, color)
	}

	// Left border
	if border.Left > 0 {
		list.PushFillRect(layout.Rect{
			X: rect.X,
			Y: rect.Y,
			W: border.Left,
			H: rect.H,
		}, color)
	}
}

// PaintBackground paints the viewport background
func PaintBackground(list *PaintList, width, height float32, color css.Color) {
	list.PushFillRect(layout.Rect{
		X: 0,
		Y: 0,
		W: width,
		H: height,
	}, color)
}
