package paint

import (
	"image"
	"image/color"
	"image/png"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// Rasterize converts paint operations to an image
func Rasterize(list *PaintList, width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for _, op := range list.Ops {
		switch op.Kind {
		case OpFillRect:
			fillRect(img, op)
		case OpStrokeRect:
			strokeRect(img, op)
		case OpDrawText:
			drawText(img, op)
		case OpClipRect:
			// TODO: implement clipping
		}
	}

	return img
}

// SavePNG saves the image to a PNG file
func SavePNG(img *image.RGBA, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}

func fillRect(img *image.RGBA, op PaintOp) {
	col := color.RGBA{op.Color.R, op.Color.G, op.Color.B, op.Color.A}

	x0 := int(op.Rect.X)
	y0 := int(op.Rect.Y)
	x1 := int(op.Rect.X + op.Rect.W)
	y1 := int(op.Rect.Y + op.Rect.H)

	bounds := img.Bounds()
	if x0 < bounds.Min.X {
		x0 = bounds.Min.X
	}
	if y0 < bounds.Min.Y {
		y0 = bounds.Min.Y
	}
	if x1 > bounds.Max.X {
		x1 = bounds.Max.X
	}
	if y1 > bounds.Max.Y {
		y1 = bounds.Max.Y
	}

	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			img.Set(x, y, col)
		}
	}
}

func strokeRect(img *image.RGBA, op PaintOp) {
	col := color.RGBA{op.Color.R, op.Color.G, op.Color.B, op.Color.A}

	x0 := int(op.Rect.X)
	y0 := int(op.Rect.Y)
	x1 := int(op.Rect.X + op.Rect.W)
	y1 := int(op.Rect.Y + op.Rect.H)

	// Top edge
	for x := x0; x < x1; x++ {
		img.Set(x, y0, col)
	}
	// Bottom edge
	for x := x0; x < x1; x++ {
		img.Set(x, y1-1, col)
	}
	// Left edge
	for y := y0; y < y1; y++ {
		img.Set(x0, y, col)
	}
	// Right edge
	for y := y0; y < y1; y++ {
		img.Set(x1-1, y, col)
	}
}

func drawText(img *image.RGBA, op PaintOp) {
	face := basicfont.Face7x13
	col := color.RGBA{op.Color.R, op.Color.G, op.Color.B, op.Color.A}

	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: face,
	}

	// Position text with baseline offset
	x := int(op.Rect.X)
	y := int(op.Rect.Y + op.FontSize) // Approximate baseline

	drawer.Dot = fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}
	drawer.DrawString(op.Text)
}
