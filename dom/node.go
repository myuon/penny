package dom

type NodeID int

const InvalidNodeID NodeID = -1

type NodeType int

const (
	NodeTypeElement NodeType = iota
	NodeTypeText
)

type Node struct {
	ID       NodeID
	Type     NodeType
	Tag      string            // element
	Attr     map[string]string // element
	Text     string            // text
	Parent   NodeID
	Children []NodeID
}

type DOM struct {
	Nodes []Node
	Root  NodeID
}

func NewDOM() *DOM {
	return &DOM{
		Nodes: []Node{},
		Root:  InvalidNodeID,
	}
}

func (d *DOM) CreateElement(tag string) NodeID {
	id := NodeID(len(d.Nodes))
	d.Nodes = append(d.Nodes, Node{
		ID:       id,
		Type:     NodeTypeElement,
		Tag:      tag,
		Attr:     make(map[string]string),
		Parent:   InvalidNodeID,
		Children: []NodeID{},
	})
	return id
}

func (d *DOM) CreateText(text string) NodeID {
	id := NodeID(len(d.Nodes))
	d.Nodes = append(d.Nodes, Node{
		ID:       id,
		Type:     NodeTypeText,
		Text:     text,
		Parent:   InvalidNodeID,
		Children: []NodeID{},
	})
	return id
}

func (d *DOM) AppendChild(parent, child NodeID) {
	d.Nodes[parent].Children = append(d.Nodes[parent].Children, child)
	d.Nodes[child].Parent = parent
}

func (d *DOM) SetAttribute(nodeID NodeID, key, value string) {
	d.Nodes[nodeID].Attr[key] = value
}

func (d *DOM) GetNode(id NodeID) *Node {
	if id < 0 || int(id) >= len(d.Nodes) {
		return nil
	}
	return &d.Nodes[id]
}
