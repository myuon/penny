package layout

import (
	"github.com/myuon/penny/css"
	"github.com/myuon/penny/dom"
)

type LayoutNodeID int32

const InvalidLayoutNodeID LayoutNodeID = -1

type Rect struct {
	X, Y, W, H float32
}

type LayoutNode struct {
	ID       LayoutNodeID
	DomNode  dom.NodeID
	Style    css.Style
	Children []LayoutNodeID
	Rect     Rect
	Text     string // for text nodes
}

type LayoutTree struct {
	Nodes []LayoutNode
	Root  LayoutNodeID
}

func NewLayoutTree() *LayoutTree {
	return &LayoutTree{
		Nodes: []LayoutNode{},
		Root:  InvalidLayoutNodeID,
	}
}

func (t *LayoutTree) CreateNode(domNode dom.NodeID, style css.Style) LayoutNodeID {
	id := LayoutNodeID(len(t.Nodes))
	t.Nodes = append(t.Nodes, LayoutNode{
		ID:       id,
		DomNode:  domNode,
		Style:    style,
		Children: []LayoutNodeID{},
		Rect:     Rect{},
	})
	return id
}

func (t *LayoutTree) AppendChild(parent, child LayoutNodeID) {
	t.Nodes[parent].Children = append(t.Nodes[parent].Children, child)
}

func (t *LayoutTree) GetNode(id LayoutNodeID) *LayoutNode {
	if id < 0 || int(id) >= len(t.Nodes) {
		return nil
	}
	return &t.Nodes[id]
}
