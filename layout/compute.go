package layout

// ComputeLayout calculates the geometry (x, y, w, h) for all nodes
func ComputeLayout(tree *LayoutTree, viewportWidth, viewportHeight float32) {
	if tree.Root == InvalidLayoutNodeID {
		return
	}

	// Start layout from root
	root := tree.GetNode(tree.Root)
	if root == nil {
		return
	}

	// Root gets full viewport
	root.Rect.X = 0
	root.Rect.Y = 0
	root.Rect.W = viewportWidth
	root.Rect.H = viewportHeight

	// Layout children
	layoutChildren(tree, tree.Root)
}

func layoutChildren(tree *LayoutTree, nodeID LayoutNodeID) {
	node := tree.GetNode(nodeID)
	if node == nil {
		return
	}

	// Calculate content area (after padding/margin)
	contentX := node.Rect.X + node.Style.Margin.Left + node.Style.Padding.Left
	contentY := node.Rect.Y + node.Style.Margin.Top + node.Style.Padding.Top
	contentW := node.Rect.W - node.Style.Margin.Left - node.Style.Margin.Right -
		node.Style.Padding.Left - node.Style.Padding.Right

	// Track current Y position for block layout
	currentY := contentY

	for _, childID := range node.Children {
		child := tree.GetNode(childID)
		if child == nil {
			continue
		}

		// Calculate child dimensions
		childW := contentW
		if child.Style.Width != nil {
			childW = *child.Style.Width
		}

		childH := estimateHeight(tree, childID)
		if child.Style.Height != nil {
			childH = *child.Style.Height
		}

		// Position child
		child.Rect.X = contentX + child.Style.Margin.Left
		child.Rect.Y = currentY + child.Style.Margin.Top
		child.Rect.W = childW - child.Style.Margin.Left - child.Style.Margin.Right
		child.Rect.H = childH

		// Move Y for next sibling (block layout)
		currentY = child.Rect.Y + child.Rect.H + child.Style.Margin.Bottom

		// Recursively layout grandchildren
		layoutChildren(tree, childID)
	}

	// Update parent height if auto
	if node.Style.Height == nil && len(node.Children) > 0 {
		lastChild := tree.GetNode(node.Children[len(node.Children)-1])
		if lastChild != nil {
			newH := (lastChild.Rect.Y + lastChild.Rect.H + lastChild.Style.Margin.Bottom) -
				node.Rect.Y + node.Style.Padding.Bottom + node.Style.Margin.Bottom
			if newH > node.Rect.H {
				node.Rect.H = newH
			}
		}
	}
}

func estimateHeight(tree *LayoutTree, nodeID LayoutNodeID) float32 {
	node := tree.GetNode(nodeID)
	if node == nil {
		return 0
	}

	// Text node: estimate based on font size
	if node.Text != "" {
		lineHeight := node.Style.FontSize * 1.5
		return lineHeight + node.Style.Padding.Top + node.Style.Padding.Bottom
	}

	// Element with explicit height
	if node.Style.Height != nil {
		return *node.Style.Height
	}

	// Sum children heights
	var totalH float32
	for _, childID := range node.Children {
		child := tree.GetNode(childID)
		if child != nil {
			totalH += estimateHeight(tree, childID)
			totalH += child.Style.Margin.Top + child.Style.Margin.Bottom
		}
	}

	return totalH + node.Style.Padding.Top + node.Style.Padding.Bottom
}
