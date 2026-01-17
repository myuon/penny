package paint

import (
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
