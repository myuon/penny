package css

type Display uint8

const (
	DisplayBlock Display = iota
	DisplayInline
	DisplayNone
	DisplayFlex
)

func (d Display) String() string {
	switch d {
	case DisplayBlock:
		return "block"
	case DisplayInline:
		return "inline"
	case DisplayNone:
		return "none"
	case DisplayFlex:
		return "flex"
	default:
		return "unknown"
	}
}

type JustifyContent uint8

const (
	JustifyFlexStart JustifyContent = iota
	JustifyFlexEnd
	JustifyCenter
	JustifySpaceBetween
	JustifySpaceAround
)

type AlignItems uint8

const (
	AlignFlexStart AlignItems = iota
	AlignFlexEnd
	AlignCenter
	AlignStretch
)

type Color struct {
	R, G, B, A uint8
}

var (
	ColorBlack       = Color{0, 0, 0, 255}
	ColorWhite       = Color{255, 255, 255, 255}
	ColorTransparent = Color{0, 0, 0, 0}
)

type Edges struct {
	Top, Right, Bottom, Left float32
}

type Style struct {
	Display        Display
	Width, Height  *float32 // nil = auto
	Margin         Edges
	Padding        Edges
	Border         Edges
	Background     Color
	BorderColor    Color
	FontSize       float32
	Color          Color
	FlexGrow       float32
	JustifyContent JustifyContent
	AlignItems     AlignItems
}

func DefaultStyle() Style {
	return Style{
		Display:        DisplayBlock,
		Width:          nil,
		Height:         nil,
		Margin:         Edges{},
		Padding:        Edges{},
		Border:         Edges{},
		Background:     ColorTransparent,
		BorderColor:    ColorBlack,
		FontSize:       16,
		Color:          ColorBlack,
		FlexGrow:       0,
		JustifyContent: JustifyFlexStart,
		AlignItems:     AlignStretch,
	}
}
