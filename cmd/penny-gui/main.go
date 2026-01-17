package main

import (
	"fmt"
	"image"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/op/clip"
	giopaint "gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/myuon/penny/css"
	"github.com/myuon/penny/dom"
	"github.com/myuon/penny/layout"
	"github.com/myuon/penny/paint"
)

const (
	windowWidth  = 800
	windowHeight = 600
)

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

	img := render(document, stylesheet)

	go func() {
		w := new(app.Window)
		w.Option(
			app.Title("Penny Browser - "+input),
			app.Size(unit.Dp(windowWidth), unit.Dp(windowHeight)),
		)

		if err := run(w, img); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}()

	app.Main()
}

func render(document *dom.DOM, stylesheet *css.Stylesheet) *image.RGBA {
	layoutTree := layout.BuildLayoutTree(document, stylesheet)
	layout.ComputeLayout(layoutTree, windowWidth, windowHeight)

	paintList := paint.NewPaintList()
	paint.PaintBackground(paintList, windowWidth, windowHeight, css.ColorWhite)
	ops := paint.Paint(layoutTree)
	paintList.Ops = append(paintList.Ops, ops.Ops...)

	return paint.Rasterize(paintList, windowWidth, windowHeight)
}

func run(w *app.Window, img *image.RGBA) error {
	var ops op.Ops
	imgOp := giopaint.NewImageOp(img)

	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			ops.Reset()

			imgOp.Add(&ops)
			stack := clip.Rect{Max: image.Pt(windowWidth, windowHeight)}.Push(&ops)
			giopaint.PaintOp{}.Add(&ops)
			stack.Pop()

			e.Frame(&ops)
		}
	}
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
