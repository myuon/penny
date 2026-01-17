package dom

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

func Parse(r io.Reader) (*DOM, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	dom := NewDOM()
	rootID := buildDOM(dom, doc, InvalidNodeID)
	dom.Root = rootID

	return dom, nil
}

func ParseString(s string) (*DOM, error) {
	return Parse(strings.NewReader(s))
}

func buildDOM(dom *DOM, n *html.Node, parentID NodeID) NodeID {
	var nodeID NodeID

	switch n.Type {
	case html.ElementNode:
		nodeID = dom.CreateElement(n.Data)
		for _, attr := range n.Attr {
			dom.SetAttribute(nodeID, attr.Key, attr.Val)
		}
	case html.TextNode:
		text := strings.TrimSpace(n.Data)
		if text == "" {
			// Skip empty text nodes
			nodeID = InvalidNodeID
		} else {
			nodeID = dom.CreateText(text)
		}
	case html.DocumentNode:
		// Document node: process children and return first valid child
		var firstChild NodeID = InvalidNodeID
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			childID := buildDOM(dom, c, parentID)
			if childID != InvalidNodeID && firstChild == InvalidNodeID {
				firstChild = childID
			}
		}
		return firstChild
	default:
		// Skip other node types (comments, doctype, etc.)
		return InvalidNodeID
	}

	if nodeID == InvalidNodeID {
		return InvalidNodeID
	}

	if parentID != InvalidNodeID {
		dom.AppendChild(parentID, nodeID)
	}

	// Process children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		buildDOM(dom, c, nodeID)
	}

	return nodeID
}
