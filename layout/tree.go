package layout

import (
	"fmt"

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

func (t *LayoutTree) Dump() string {
	var result string
	t.dumpNode(t.Root, 0, &result)
	return result
}

func (t *LayoutTree) dumpNode(id LayoutNodeID, indent int, result *string) {
	node := t.GetNode(id)
	if node == nil {
		return
	}

	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	rect := fmt.Sprintf("(%.1f, %.1f, %.1f, %.1f)", node.Rect.X, node.Rect.Y, node.Rect.W, node.Rect.H)
	if node.Text != "" {
		*result += fmt.Sprintf("%s[text] %s \"%s\"\n", prefix, rect, node.Text)
	} else {
		*result += fmt.Sprintf("%s[%d] %s display=%s\n", prefix, node.DomNode, rect, node.Style.Display)
	}

	for _, childID := range node.Children {
		t.dumpNode(childID, indent+1, result)
	}
}
