package main

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	giopaint "gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/myuon/penny/css"
	"github.com/myuon/penny/dom"
	pennylayout "github.com/myuon/penny/layout"
	"github.com/myuon/penny/paint"
)

const (
	contentWidth  = 800
	contentHeight = 600
	devToolsWidth = 400
	windowWidth   = contentWidth + devToolsWidth
	windowHeight  = 600
)

type DevTab int

const (
	TabDOM DevTab = iota
	TabStylesheet
	TabLayoutTree
	TabPaintOps
)

type Browser struct {
	document   *dom.DOM
	stylesheet *css.Stylesheet
	layoutTree *pennylayout.LayoutTree
	paintList  *paint.PaintList
	canvas     *image.RGBA

	// UI state
	activeTab   DevTab
	btnDOM      widget.Clickable
	btnStyle    widget.Clickable
	btnLayout   widget.Clickable
	btnPaint    widget.Clickable
	devScroll   widget.List
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: penny-gui <URL or file>")
		os.Exit(1)
	}

	input := os.Args[1]

	var htmlContent string
	var baseURL *url.URL
	var baseDir string

	if isURL(input) {
		fmt.Printf("Fetching: %s\n", input)
		content, err := fetchURL(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to fetch URL: %v\n", err)
			os.Exit(1)
		}
		htmlContent = content
		baseURL, _ = url.Parse(input)
	} else {
		data, err := os.ReadFile(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read file: %v\n", err)
			os.Exit(1)
		}
		htmlContent = string(data)
		baseDir = filepath.Dir(input)
	}

	document, err := dom.ParseString(htmlContent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse HTML: %v\n", err)
		os.Exit(1)
	}

	var stylesheet *css.Stylesheet
	if baseURL != nil {
		stylesheet = loadStylesheetsFromURL(document, baseURL)
	} else {
		stylesheet = loadStylesheetsFromDir(document, baseDir)
	}

	browser := &Browser{
		document:   document,
		stylesheet: stylesheet,
		activeTab:  TabDOM,
	}
	browser.devScroll.Axis = layout.Vertical
	browser.render()

	go func() {
		w := new(app.Window)
		w.Option(
			app.Title("Penny Browser - "+input),
			app.Size(unit.Dp(windowWidth), unit.Dp(windowHeight)),
		)

		if err := browser.run(w); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}()

	app.Main()
}

func (b *Browser) render() {
	b.layoutTree = pennylayout.BuildLayoutTree(b.document, b.stylesheet)
	pennylayout.ComputeLayout(b.layoutTree, contentWidth, contentHeight)

	b.paintList = paint.NewPaintList()
	paint.PaintBackground(b.paintList, contentWidth, contentHeight, css.ColorWhite)
	ops := paint.Paint(b.layoutTree)
	b.paintList.Ops = append(b.paintList.Ops, ops.Ops...)

	b.canvas = paint.Rasterize(b.paintList, contentWidth, contentHeight)
}

func (b *Browser) run(w *app.Window) error {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	var ops op.Ops

	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			// Handle button clicks
			if b.btnDOM.Clicked(gtx) {
				b.activeTab = TabDOM
			}
			if b.btnStyle.Clicked(gtx) {
				b.activeTab = TabStylesheet
			}
			if b.btnLayout.Clicked(gtx) {
				b.activeTab = TabLayoutTree
			}
			if b.btnPaint.Clicked(gtx) {
				b.activeTab = TabPaintOps
			}

			b.layout(gtx, th)
			e.Frame(gtx.Ops)
		}
	}
}

func (b *Browser) layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{}.Layout(gtx,
		// Content area (left)
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return b.layoutContent(gtx)
		}),
		// DevTools area (right)
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return b.layoutDevTools(gtx, th)
		}),
	)
}

func (b *Browser) layoutContent(gtx layout.Context) layout.Dimensions {
	imgOp := giopaint.NewImageOp(b.canvas)
	imgOp.Add(gtx.Ops)
	stack := clip.Rect{Max: image.Pt(contentWidth, contentHeight)}.Push(gtx.Ops)
	giopaint.PaintOp{}.Add(gtx.Ops)
	stack.Pop()

	return layout.Dimensions{Size: image.Pt(contentWidth, contentHeight)}
}

func (b *Browser) layoutDevTools(gtx layout.Context, th *material.Theme) layout.Dimensions {
	// Background
	bgColor := color.NRGBA{R: 40, G: 40, B: 40, A: 255}
	stack := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
	giopaint.ColorOp{Color: bgColor}.Add(gtx.Ops)
	giopaint.PaintOp{}.Add(gtx.Ops)
	stack.Pop()

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Tab buttons
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return b.tabButton(gtx, th, &b.btnDOM, "DOM", TabDOM)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return b.tabButton(gtx, th, &b.btnStyle, "Style", TabStylesheet)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return b.tabButton(gtx, th, &b.btnLayout, "Layout", TabLayoutTree)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return b.tabButton(gtx, th, &b.btnPaint, "Paint", TabPaintOps)
				}),
			)
		}),
		// Content area
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return b.layoutDevContent(gtx, th)
		}),
	)
}

func (b *Browser) tabButton(gtx layout.Context, th *material.Theme, btn *widget.Clickable, label string, tab DevTab) layout.Dimensions {
	var bgColor color.NRGBA
	if b.activeTab == tab {
		bgColor = color.NRGBA{R: 60, G: 60, B: 60, A: 255}
	} else {
		bgColor = color.NRGBA{R: 50, G: 50, B: 50, A: 255}
	}

	return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		btnStyle := material.Button(th, btn, label)
		btnStyle.Background = bgColor
		btnStyle.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		return btnStyle.Layout(gtx)
	})
}

func (b *Browser) layoutDevContent(gtx layout.Context, th *material.Theme) layout.Dimensions {
	var content string
	switch b.activeTab {
	case TabDOM:
		content = b.document.Dump()
	case TabStylesheet:
		if b.stylesheet != nil {
			content = b.stylesheet.Dump()
		} else {
			content = "(no stylesheet)"
		}
	case TabLayoutTree:
		content = b.layoutTree.Dump()
	case TabPaintOps:
		content = b.paintList.Dump()
	}

	return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return material.List(th, &b.devScroll).Layout(gtx, 1, func(gtx layout.Context, _ int) layout.Dimensions {
			lbl := material.Body1(th, content)
			lbl.Color = color.NRGBA{R: 200, G: 200, B: 200, A: 255}
			return lbl.Layout(gtx)
		})
	})
}

func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

func fetchURL(urlStr string) (string, error) {
	resp, err := http.Get(urlStr)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func loadStylesheetsFromDir(d *dom.DOM, baseDir string) *css.Stylesheet {
	var allRules []css.Rule

	var walk func(nodeID dom.NodeID)
	walk = func(nodeID dom.NodeID) {
		node := d.GetNode(nodeID)
		if node == nil {
			return
		}

		if node.Type == dom.NodeTypeElement && node.Tag == "link" {
			rel, hasRel := node.Attr["rel"]
			href, hasHref := node.Attr["href"]
			if hasRel && rel == "stylesheet" && hasHref {
				cssPath := filepath.Join(baseDir, href)
				if data, err := os.ReadFile(cssPath); err == nil {
					if sheet, err := css.Parse(string(data)); err == nil {
						allRules = append(allRules, sheet.Rules...)
						fmt.Printf("Loaded CSS: %s\n", cssPath)
					}
				}
			}
		}

		if node.Type == dom.NodeTypeElement && node.Tag == "style" {
			cssText := extractTextContent(d, nodeID)
			if cssText != "" {
				if sheet, err := css.Parse(cssText); err == nil {
					allRules = append(allRules, sheet.Rules...)
					fmt.Println("Loaded CSS: <style>")
				}
			}
		}

		for _, childID := range node.Children {
			walk(childID)
		}
	}

	walk(d.Root)

	if len(allRules) == 0 {
		return nil
	}

	return &css.Stylesheet{Rules: allRules}
}

func loadStylesheetsFromURL(d *dom.DOM, baseURL *url.URL) *css.Stylesheet {
	var allRules []css.Rule

	var walk func(nodeID dom.NodeID)
	walk = func(nodeID dom.NodeID) {
		node := d.GetNode(nodeID)
		if node == nil {
			return
		}

		if node.Type == dom.NodeTypeElement && node.Tag == "link" {
			rel, hasRel := node.Attr["rel"]
			href, hasHref := node.Attr["href"]
			if hasRel && rel == "stylesheet" && hasHref {
				cssURL := resolveURL(baseURL, href)
				if content, err := fetchURL(cssURL); err == nil {
					if sheet, err := css.Parse(content); err == nil {
						allRules = append(allRules, sheet.Rules...)
						fmt.Printf("Loaded CSS: %s\n", cssURL)
					}
				}
			}
		}

		if node.Type == dom.NodeTypeElement && node.Tag == "style" {
			cssText := extractTextContent(d, nodeID)
			if cssText != "" {
				if sheet, err := css.Parse(cssText); err == nil {
					allRules = append(allRules, sheet.Rules...)
					fmt.Println("Loaded CSS: <style>")
				}
			}
		}

		for _, childID := range node.Children {
			walk(childID)
		}
	}

	walk(d.Root)

	if len(allRules) == 0 {
		return nil
	}

	return &css.Stylesheet{Rules: allRules}
}

func resolveURL(base *url.URL, ref string) string {
	refURL, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	return base.ResolveReference(refURL).String()
}

func extractTextContent(d *dom.DOM, nodeID dom.NodeID) string {
	var text string
	var walk func(id dom.NodeID)
	walk = func(id dom.NodeID) {
		node := d.GetNode(id)
		if node == nil {
			return
		}
		if node.Type == dom.NodeTypeText {
			text += node.Text
		}
		for _, childID := range node.Children {
			walk(childID)
		}
	}
	walk(nodeID)
	return text
}
