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

	// Root should be auto-inserted <html>
	root := dom.GetNode(dom.Root)
	if root == nil {
		t.Fatal("root is nil")
	}
	if root.Tag != "html" {
		t.Errorf("expected root tag 'html', got %q", root.Tag)
	}

	// <html> should have <body> as child
	if len(root.Children) != 1 {
		t.Errorf("expected 1 child (body), got %d", len(root.Children))
	}

	bodyNode := dom.GetNode(root.Children[0])
	if bodyNode.Tag != "body" {
		t.Errorf("expected 'body', got %q", bodyNode.Tag)
	}

	// <body> should have <p> as child
	if len(bodyNode.Children) != 1 {
		t.Errorf("expected 1 child (p), got %d", len(bodyNode.Children))
	}

	pNode := dom.GetNode(bodyNode.Children[0])
	if pNode.Tag != "p" {
		t.Errorf("expected 'p', got %q", pNode.Tag)
	}

	// <p> should have text node as child
	if len(pNode.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(pNode.Children))
	}

	textNode := dom.GetNode(pNode.Children[0])
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

	// Root should be auto-inserted <html>
	root := dom.GetNode(dom.Root)
	if root == nil {
		t.Fatal("root is nil")
	}
	if root.Tag != "html" {
		t.Errorf("expected root tag 'html', got %q", root.Tag)
	}

	// Find <body>
	bodyNode := dom.GetNode(root.Children[0])
	if bodyNode.Tag != "body" {
		t.Errorf("expected 'body', got %q", bodyNode.Tag)
	}

	// Find <div> inside body
	divNode := dom.GetNode(bodyNode.Children[0])
	if divNode.Tag != "div" {
		t.Errorf("expected 'div', got %q", divNode.Tag)
	}

	// Check class attribute
	if class, ok := divNode.Attr["class"]; !ok || class != "container" {
		t.Errorf("expected class='container', got %q", divNode.Attr["class"])
	}

	// Should have 2 children (h1 and p)
	if len(divNode.Children) != 2 {
		t.Errorf("expected 2 children, got %d", len(divNode.Children))
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

	// Root should be auto-inserted <html>
	root := dom.GetNode(dom.Root)
	if root == nil {
		t.Fatal("root is nil")
	}
	if root.Tag != "html" {
		t.Errorf("expected root tag 'html', got %q", root.Tag)
	}

	// Find <body>
	bodyNode := dom.GetNode(root.Children[0])
	if bodyNode.Tag != "body" {
		t.Errorf("expected 'body', got %q", bodyNode.Tag)
	}

	// Body should have 3 <p> children
	if len(bodyNode.Children) != 3 {
		t.Errorf("expected 3 children, got %d", len(bodyNode.Children))
	}

	expectedTexts := []string{"First", "Second", "Third"}
	for i, childID := range bodyNode.Children {
		child := dom.GetNode(childID)
		if child.Tag != "p" {
			t.Errorf("expected 'p', got %q", child.Tag)
		}
		if len(child.Children) > 0 {
			textNode := dom.GetNode(child.Children[0])
			if textNode.Text != expectedTexts[i] {
				t.Errorf("expected %q, got %q", expectedTexts[i], textNode.Text)
			}
		}
	}

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

	// Root should be auto-inserted <html>
	root := dom.GetNode(dom.Root)
	if root == nil {
		t.Fatal("root is nil")
	}
	if root.Tag != "html" {
		t.Errorf("expected root tag 'html', got %q", root.Tag)
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

	// Root should be auto-inserted <html>
	root := dom.GetNode(dom.Root)
	if root == nil {
		t.Fatal("root is nil")
	}
	if root.Tag != "html" {
		t.Errorf("expected root tag 'html', got %q", root.Tag)
	}

	// Find the <div> inside body
	bodyNode := dom.GetNode(root.Children[0])
	divNode := dom.GetNode(bodyNode.Children[0])
	if divNode.Tag != "div" {
		t.Errorf("expected 'div', got %q", divNode.Tag)
	}

	// Should have 4 void element children
	if len(divNode.Children) != 4 {
		t.Errorf("expected 4 children, got %d", len(divNode.Children))
	}

	// Check each void element
	expectedTags := []string{"br", "hr", "img", "input"}
	for i, childID := range divNode.Children {
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

	// Root should be auto-inserted <html>
	root := dom.GetNode(dom.Root)
	if root == nil {
		t.Fatal("root is nil")
	}
	if root.Tag != "html" {
		t.Errorf("expected 'html', got %q", root.Tag)
	}

	// Find <p> inside body
	bodyNode := dom.GetNode(root.Children[0])
	pNode := dom.GetNode(bodyNode.Children[0])
	if pNode.Tag != "p" {
		t.Errorf("expected 'p', got %q", pNode.Tag)
	}

	// Should have 3 children: "Hello ", <strong>, "!"
	if len(pNode.Children) != 3 {
		t.Errorf("expected 3 children, got %d", len(pNode.Children))
	}

	t.Logf("DOM:\n%s", dom.Dump())
}

func TestParseLinkOnly(t *testing.T) {
	// Link tag without html/head wrappers
	input := `<link rel="stylesheet" href="style.css">`

	dom, err := ParseString(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Root should be auto-inserted <html>
	root := dom.GetNode(dom.Root)
	if root == nil {
		t.Fatal("root is nil")
	}
	if root.Tag != "html" {
		t.Errorf("expected root tag 'html', got %q", root.Tag)
	}

	// <html> should have <head> as child
	if len(root.Children) != 1 {
		t.Errorf("expected 1 child (head), got %d", len(root.Children))
	}

	headNode := dom.GetNode(root.Children[0])
	if headNode.Tag != "head" {
		t.Errorf("expected 'head', got %q", headNode.Tag)
	}

	// <head> should have <link> as child
	if len(headNode.Children) != 1 {
		t.Errorf("expected 1 child (link), got %d", len(headNode.Children))
	}

	linkNode := dom.GetNode(headNode.Children[0])
	if linkNode.Tag != "link" {
		t.Errorf("expected 'link', got %q", linkNode.Tag)
	}

	// Check attributes
	if rel, ok := linkNode.Attr["rel"]; !ok || rel != "stylesheet" {
		t.Errorf("expected rel='stylesheet', got %q", linkNode.Attr["rel"])
	}
	if href, ok := linkNode.Attr["href"]; !ok || href != "style.css" {
		t.Errorf("expected href='style.css', got %q", linkNode.Attr["href"])
	}

	t.Logf("DOM:\n%s", dom.Dump())
}

func TestParseLinkThenBody(t *testing.T) {
	// Link tag followed by body content
	input := `<link rel="stylesheet" href="style.css">
<p>Hello World</p>`

	dom, err := ParseString(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Root should be auto-inserted <html>
	root := dom.GetNode(dom.Root)
	if root == nil {
		t.Fatal("root is nil")
	}
	if root.Tag != "html" {
		t.Errorf("expected root tag 'html', got %q", root.Tag)
	}

	// <html> should have 2 children: <head> and <body>
	if len(root.Children) != 2 {
		t.Errorf("expected 2 children (head, body), got %d", len(root.Children))
	}

	// First child should be <head>
	headNode := dom.GetNode(root.Children[0])
	if headNode.Tag != "head" {
		t.Errorf("expected 'head', got %q", headNode.Tag)
	}

	// <head> should have <link>
	if len(headNode.Children) != 1 {
		t.Errorf("expected 1 child in head, got %d", len(headNode.Children))
	}
	linkNode := dom.GetNode(headNode.Children[0])
	if linkNode.Tag != "link" {
		t.Errorf("expected 'link', got %q", linkNode.Tag)
	}

	// Second child should be <body>
	bodyNode := dom.GetNode(root.Children[1])
	if bodyNode.Tag != "body" {
		t.Errorf("expected 'body', got %q", bodyNode.Tag)
	}

	// <body> should have <p>
	if len(bodyNode.Children) != 1 {
		t.Errorf("expected 1 child in body, got %d", len(bodyNode.Children))
	}
	pNode := dom.GetNode(bodyNode.Children[0])
	if pNode.Tag != "p" {
		t.Errorf("expected 'p', got %q", pNode.Tag)
	}

	t.Logf("DOM:\n%s", dom.Dump())
}

func TestParseMetaAndTitle(t *testing.T) {
	// Multiple head elements
	input := `<meta charset="UTF-8">
<title>Test Page</title>
<div>Content</div>`

	dom, err := ParseString(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Root should be <html>
	root := dom.GetNode(dom.Root)
	if root.Tag != "html" {
		t.Errorf("expected 'html', got %q", root.Tag)
	}

	// Should have head and body
	if len(root.Children) != 2 {
		t.Errorf("expected 2 children, got %d", len(root.Children))
	}

	headNode := dom.GetNode(root.Children[0])
	if headNode.Tag != "head" {
		t.Errorf("expected 'head', got %q", headNode.Tag)
	}

	// Head should have meta and title
	if len(headNode.Children) != 2 {
		t.Errorf("expected 2 children in head, got %d", len(headNode.Children))
	}

	metaNode := dom.GetNode(headNode.Children[0])
	if metaNode.Tag != "meta" {
		t.Errorf("expected 'meta', got %q", metaNode.Tag)
	}

	titleNode := dom.GetNode(headNode.Children[1])
	if titleNode.Tag != "title" {
		t.Errorf("expected 'title', got %q", titleNode.Tag)
	}

	bodyNode := dom.GetNode(root.Children[1])
	if bodyNode.Tag != "body" {
		t.Errorf("expected 'body', got %q", bodyNode.Tag)
	}

	t.Logf("DOM:\n%s", dom.Dump())
}

func TestParseStyleTag(t *testing.T) {
	// Style tag with CSS content
	input := `<style>
body { color: red; }
</style>
<p>Styled</p>`

	dom, err := ParseString(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	root := dom.GetNode(dom.Root)
	if root.Tag != "html" {
		t.Errorf("expected 'html', got %q", root.Tag)
	}

	// Should have head and body
	if len(root.Children) != 2 {
		t.Errorf("expected 2 children, got %d", len(root.Children))
	}

	headNode := dom.GetNode(root.Children[0])
	if headNode.Tag != "head" {
		t.Errorf("expected 'head', got %q", headNode.Tag)
	}

	// Head should have style
	styleNode := dom.GetNode(headNode.Children[0])
	if styleNode.Tag != "style" {
		t.Errorf("expected 'style', got %q", styleNode.Tag)
	}

	// Style should have text content
	if len(styleNode.Children) != 1 {
		t.Errorf("expected 1 child in style, got %d", len(styleNode.Children))
	}

	t.Logf("DOM:\n%s", dom.Dump())
}
