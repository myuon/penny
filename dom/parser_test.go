package dom

import (
	"testing"
)

func TestParseFullHTML(t *testing.T) {
	input := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
<p>Hello</p>
</body>
</html>`

	dom, err := ParseString(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Root should be <html>
	root := dom.GetNode(dom.Root)
	if root == nil {
		t.Fatal("root is nil")
	}
	if root.Tag != "html" {
		t.Errorf("expected root tag 'html', got %q", root.Tag)
	}

	t.Logf("DOM:\n%s", dom.Dump())
}

func TestParseNoHTMLTag(t *testing.T) {
	// HTML without <html> tag, starts directly with <body>
	input := `<body>
<p>Hello</p>
</body>`

	dom, err := ParseString(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Root should be <body>
	root := dom.GetNode(dom.Root)
	if root == nil {
		t.Fatal("root is nil")
	}
	if root.Tag != "body" {
		t.Errorf("expected root tag 'body', got %q", root.Tag)
	}

	// Should have <p> as child
	if len(root.Children) == 0 {
		t.Error("expected children in body")
	}

	t.Logf("DOM:\n%s", dom.Dump())
}

func TestParseNoBodyTag(t *testing.T) {
	// HTML without <html> and <body> tags, starts directly with <p>
	input := `<p>Hello World</p>`

	dom, err := ParseString(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Root should be <p>
	root := dom.GetNode(dom.Root)
	if root == nil {
		t.Fatal("root is nil")
	}
	if root.Tag != "p" {
		t.Errorf("expected root tag 'p', got %q", root.Tag)
	}

	// Should have text node as child
	if len(root.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(root.Children))
	}

	textNode := dom.GetNode(root.Children[0])
	if textNode.Type != NodeTypeText {
		t.Errorf("expected text node, got %v", textNode.Type)
	}
	if textNode.Text != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", textNode.Text)
	}

	t.Logf("DOM:\n%s", dom.Dump())
}

func TestParseNoBodyTagWithDiv(t *testing.T) {
	// HTML without <html> and <body> tags, starts directly with <div>
	input := `<div class="container">
<h1>Title</h1>
<p>Content</p>
</div>`

	dom, err := ParseString(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Root should be <div>
	root := dom.GetNode(dom.Root)
	if root == nil {
		t.Fatal("root is nil")
	}
	if root.Tag != "div" {
		t.Errorf("expected root tag 'div', got %q", root.Tag)
	}

	// Check class attribute
	if class, ok := root.Attr["class"]; !ok || class != "container" {
		t.Errorf("expected class='container', got %q", root.Attr["class"])
	}

	// Should have 2 children (h1 and p)
	if len(root.Children) != 2 {
		t.Errorf("expected 2 children, got %d", len(root.Children))
	}

	t.Logf("DOM:\n%s", dom.Dump())
}

func TestParseMultipleTopLevelElements(t *testing.T) {
	// Multiple elements at top level (no wrapper)
	input := `<p>First</p>
<p>Second</p>
<p>Third</p>`

	dom, err := ParseString(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Root should be first <p>
	root := dom.GetNode(dom.Root)
	if root == nil {
		t.Fatal("root is nil")
	}
	if root.Tag != "p" {
		t.Errorf("expected root tag 'p', got %q", root.Tag)
	}

	// Note: Current parser only tracks first element as root
	// The other <p> elements become siblings (not children)
	// This is a limitation of the current parser

	t.Logf("DOM:\n%s", dom.Dump())
}

func TestParseNestedElements(t *testing.T) {
	input := `<div>
<div>
<div>
<p>Deep</p>
</div>
</div>
</div>`

	dom, err := ParseString(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	root := dom.GetNode(dom.Root)
	if root == nil {
		t.Fatal("root is nil")
	}
	if root.Tag != "div" {
		t.Errorf("expected root tag 'div', got %q", root.Tag)
	}

	// Traverse to find the <p> tag
	var findP func(id NodeID) *Node
	findP = func(id NodeID) *Node {
		node := dom.GetNode(id)
		if node == nil {
			return nil
		}
		if node.Tag == "p" {
			return node
		}
		for _, childID := range node.Children {
			if found := findP(childID); found != nil {
				return found
			}
		}
		return nil
	}

	p := findP(dom.Root)
	if p == nil {
		t.Error("could not find <p> element")
	} else if len(p.Children) != 1 {
		t.Errorf("expected 1 child in <p>, got %d", len(p.Children))
	}

	t.Logf("DOM:\n%s", dom.Dump())
}

func TestParseTextOnly(t *testing.T) {
	// Just text, no tags
	input := `Hello World`

	dom, err := ParseString(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Root should be InvalidNodeID since there's no parent for the text
	if dom.Root != InvalidNodeID {
		t.Errorf("expected InvalidNodeID root for text-only input, got %d", dom.Root)
	}

	t.Logf("DOM:\n%s", dom.Dump())
}

func TestParseEmptyInput(t *testing.T) {
	input := ``

	dom, err := ParseString(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if dom.Root != InvalidNodeID {
		t.Errorf("expected InvalidNodeID root for empty input, got %d", dom.Root)
	}
}

func TestParseVoidElements(t *testing.T) {
	input := `<div>
<br>
<hr>
<img src="test.png">
<input type="text">
</div>`

	dom, err := ParseString(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	root := dom.GetNode(dom.Root)
	if root == nil {
		t.Fatal("root is nil")
	}

	// Should have 4 void element children
	if len(root.Children) != 4 {
		t.Errorf("expected 4 children, got %d", len(root.Children))
	}

	// Check each void element
	expectedTags := []string{"br", "hr", "img", "input"}
	for i, childID := range root.Children {
		child := dom.GetNode(childID)
		if child.Tag != expectedTags[i] {
			t.Errorf("expected %q, got %q", expectedTags[i], child.Tag)
		}
	}

	t.Logf("DOM:\n%s", dom.Dump())
}

func TestParseMixedContent(t *testing.T) {
	input := `<p>Hello <strong>World</strong>!</p>`

	dom, err := ParseString(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	root := dom.GetNode(dom.Root)
	if root == nil {
		t.Fatal("root is nil")
	}
	if root.Tag != "p" {
		t.Errorf("expected 'p', got %q", root.Tag)
	}

	// Should have 3 children: "Hello ", <strong>, "!"
	if len(root.Children) != 3 {
		t.Errorf("expected 3 children, got %d", len(root.Children))
	}

	t.Logf("DOM:\n%s", dom.Dump())
}
