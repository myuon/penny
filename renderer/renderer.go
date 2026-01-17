package renderer

import (
	"github.com/myuon/penny/css"
	"github.com/myuon/penny/dom"
	"github.com/myuon/penny/layout"
	"github.com/myuon/penny/paint"
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

// Render executes the full rendering pipeline:
// DOM + CSS → LayoutTree → Layout → PaintOps → Rasterize → PNG
func (r *Renderer) Render(d *dom.DOM, stylesheet *css.Stylesheet, outputPath string) error {
	// 1. Build layout tree (cascade/compute styles)
	layoutTree := layout.BuildLayoutTree(d, stylesheet)

	// 2. Compute layout (geometry)
	layout.ComputeLayout(layoutTree, float32(r.Width), float32(r.Height))

	// 3. Paint (generate paint operations)
	paintList := paint.NewPaintList()

	// Background
	paint.PaintBackground(paintList, float32(r.Width), float32(r.Height), css.ColorWhite)

	// Paint layout tree
	ops := paint.Paint(layoutTree)
	paintList.Ops = append(paintList.Ops, ops.Ops...)

	// 4. Rasterize (convert to image)
	img := paint.Rasterize(paintList, r.Width, r.Height)

	// 5. Save to PNG
	return paint.SavePNG(img, outputPath)
}
