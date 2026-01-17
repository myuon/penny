package renderer

import (
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/myuon/penny/dom"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type Renderer struct {
	Width  int
	Height int
}

func New(width, height int) *Renderer {
	return &Renderer{
		Width:  width,
		Height: height,
	}
}

func (r *Renderer) Render(d *dom.DOM, outputPath string) error {
	// Create image
	img := image.NewRGBA(image.Rect(0, 0, r.Width, r.Height))

	// Fill background with white
	for y := 0; y < r.Height; y++ {
		for x := 0; x < r.Width; x++ {
			img.Set(x, y, color.White)
		}
	}

	// Find body element and extract text nodes
	texts := r.extractTexts(d)

	// Draw texts
	r.drawTexts(img, texts)

	// Save to file
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}

func (r *Renderer) extractTexts(d *dom.DOM) []string {
	var texts []string
	var inBody bool

	var walk func(nodeID dom.NodeID)
	walk = func(nodeID dom.NodeID) {
		node := d.GetNode(nodeID)
		if node == nil {
			return
		}

		if node.Type == dom.NodeTypeElement && node.Tag == "body" {
			inBody = true
		}

		if inBody && node.Type == dom.NodeTypeText {
			texts = append(texts, node.Text)
		}

		for _, childID := range node.Children {
			walk(childID)
		}

		if node.Type == dom.NodeTypeElement && node.Tag == "body" {
			inBody = false
		}
	}

	walk(d.Root)
	return texts
}

func (r *Renderer) drawTexts(img *image.RGBA, texts []string) {
	face := basicfont.Face7x13
	col := color.Black

	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: face,
	}

	x := 10
	y := 20
	lineHeight := 20

	for _, text := range texts {
		drawer.Dot = fixed.Point26_6{
			X: fixed.I(x),
			Y: fixed.I(y),
		}
		drawer.DrawString(text)
		y += lineHeight
	}
}
