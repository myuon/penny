package paint

import (
	"fmt"

	"github.com/myuon/penny/css"
	"github.com/myuon/penny/layout"
)

type PaintOpKind uint8

const (
	OpFillRect PaintOpKind = iota
	OpStrokeRect
	OpDrawText
	OpClipRect
)

func (k PaintOpKind) String() string {
	switch k {
	case OpFillRect:
		return "FillRect"
	case OpStrokeRect:
		return "StrokeRect"
	case OpDrawText:
		return "DrawText"
	case OpClipRect:
		return "ClipRect"
	default:
		return "Unknown"
	}
}

type PaintOp struct {
	Kind     PaintOpKind
	Rect     layout.Rect
	Color    css.Color
	Text     string
	FontSize float32
}

type PaintList struct {
	Ops []PaintOp
}

func NewPaintList() *PaintList {
	return &PaintList{
		Ops: []PaintOp{},
	}
}

func (p *PaintList) PushFillRect(rect layout.Rect, color css.Color) {
	p.Ops = append(p.Ops, PaintOp{
		Kind:  OpFillRect,
		Rect:  rect,
		Color: color,
	})
}

func (p *PaintList) PushStrokeRect(rect layout.Rect, color css.Color) {
	p.Ops = append(p.Ops, PaintOp{
		Kind:  OpStrokeRect,
		Rect:  rect,
		Color: color,
	})
}

func (p *PaintList) PushDrawText(rect layout.Rect, text string, color css.Color, fontSize float32) {
	p.Ops = append(p.Ops, PaintOp{
		Kind:     OpDrawText,
		Rect:     rect,
		Text:     text,
		Color:    color,
		FontSize: fontSize,
	})
}

func (p *PaintList) PushClipRect(rect layout.Rect) {
	p.Ops = append(p.Ops, PaintOp{
		Kind: OpClipRect,
		Rect: rect,
	})
}

func (p *PaintList) Dump() string {
	var result string
	for i, op := range p.Ops {
		rect := fmt.Sprintf("(%.1f, %.1f, %.1f, %.1f)", op.Rect.X, op.Rect.Y, op.Rect.W, op.Rect.H)
		color := fmt.Sprintf("rgba(%d,%d,%d,%d)", op.Color.R, op.Color.G, op.Color.B, op.Color.A)

		switch op.Kind {
		case OpFillRect:
			result += fmt.Sprintf("%d: FillRect %s %s\n", i, rect, color)
		case OpStrokeRect:
			result += fmt.Sprintf("%d: StrokeRect %s %s\n", i, rect, color)
		case OpDrawText:
			result += fmt.Sprintf("%d: DrawText %s %s fontSize=%.1f \"%s\"\n", i, rect, color, op.FontSize, op.Text)
		case OpClipRect:
			result += fmt.Sprintf("%d: ClipRect %s\n", i, rect)
		}
	}
	return result
}
