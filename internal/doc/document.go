package doc

import (
	"image"
	"image/color"
	"image/draw"
)

type Tool int

const (
	ToolPen Tool = iota
	ToolEraser
	ToolFill
	ToolPipette
)

func (t Tool) String() string {
	switch t {
	case ToolPen:
		return "pen"
	case ToolEraser:
		return "eraser"
	case ToolFill:
		return "fill"
	case ToolPipette:
		return "pipette"
	default:
		return "pen"
	}
}

const (
	DefaultWidth  = 64
	DefaultHeight = 64
	MinZoom       = 1
	MaxZoom       = 40
	MaxHistory    = 30
)

var Transparent = color.RGBA{0, 0, 0, 0}

type Document struct {
	Width, Height int
	Img           *image.RGBA
	Zoom          int
	Brush         color.RGBA
	Tool          Tool
	BrushSize     int
	ShowGrid      bool
	GridGap       int
	Symmetry      bool
	LockPalette   bool
	FilePath      string

	history    []*image.RGBA
	redo       []*image.RGBA
	maxHistory int
}

func New(width, height int) *Document {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	d := &Document{
		Width:      width,
		Height:     height,
		Zoom:       10,
		Brush:      color.RGBA{0, 0, 0, 255},
		Tool:       ToolPen,
		BrushSize:  1,
		ShowGrid:   true,
		GridGap:    1,
		maxHistory: MaxHistory,
	}
	d.Img = image.NewRGBA(image.Rect(0, 0, width, height))
	d.SaveState()
	return d
}

func cloneRGBA(src *image.RGBA) *image.RGBA {
	dst := image.NewRGBA(src.Bounds())
	copy(dst.Pix, src.Pix)
	return dst
}

func ToRGBA(src image.Image) *image.RGBA {
	if src == nil {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}
	if r, ok := src.(*image.RGBA); ok {
		return cloneRGBA(r)
	}
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
	return dst
}

func (d *Document) In(x, y int) bool {
	return x >= 0 && y >= 0 && x < d.Width && y < d.Height
}

func (d *Document) At(x, y int) color.RGBA {
	if !d.In(x, y) {
		return Transparent
	}
	return d.Img.RGBAAt(x, y)
}

func (d *Document) set(x, y int, c color.RGBA) {
	if !d.In(x, y) {
		return
	}
	d.Img.SetRGBA(x, y, c)
}

func (d *Document) setImage(img *image.RGBA) {
	d.Img = img
	b := img.Bounds()
	d.Width, d.Height = b.Dx(), b.Dy()
}

func (d *Document) SaveState() {
	d.history = append(d.history, cloneRGBA(d.Img))
	if len(d.history) > d.maxHistory {
		d.history = d.history[1:]
	}
	d.redo = d.redo[:0]
}

func (d *Document) CanUndo() bool { return len(d.history) > 0 }
func (d *Document) CanRedo() bool { return len(d.redo) > 0 }

func (d *Document) Undo() bool {
	if len(d.history) == 0 {
		return false
	}
	d.redo = append(d.redo, cloneRGBA(d.Img))
	last := d.history[len(d.history)-1]
	d.history = d.history[:len(d.history)-1]
	d.setImage(cloneRGBA(last))
	return true
}

func (d *Document) Redo() bool {
	if len(d.redo) == 0 {
		return false
	}
	d.history = append(d.history, cloneRGBA(d.Img))
	last := d.redo[len(d.redo)-1]
	d.redo = d.redo[:len(d.redo)-1]
	d.setImage(cloneRGBA(last))
	return true
}

func (d *Document) Clear() {
	d.SaveState()
	d.Img = image.NewRGBA(image.Rect(0, 0, d.Width, d.Height))
}

func (d *Document) Reset(width, height int) {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	d.SaveState()
	d.Width, d.Height = width, height
	d.Img = image.NewRGBA(image.Rect(0, 0, width, height))
	d.FilePath = ""
}

func (d *Document) LoadImage(img image.Image, path string) {
	d.SaveState()
	d.setImage(ToRGBA(img))
	d.FilePath = path
	if d.Width > 100 {
		d.Zoom = 4
	}
}

func (d *Document) ChangeZoom(delta int) {
	z := d.Zoom + delta
	if z < MinZoom {
		z = MinZoom
	}
	if z > MaxZoom {
		z = MaxZoom
	}
	d.Zoom = z
}

func (d *Document) SetZoom(z int) {
	if z < MinZoom {
		z = MinZoom
	}
	if z > MaxZoom {
		z = MaxZoom
	}
	d.Zoom = z
}

func (d *Document) stamp(x, y int, c color.RGBA) {
	if d.BrushSize <= 1 {
		d.set(x, y, c)
		return
	}
	r := d.BrushSize / 2
	for yy := y - r; yy <= y+r; yy++ {
		for xx := x - r; xx <= x+r; xx++ {
			d.set(xx, yy, c)
		}
	}
}

func (d *Document) DrawBrush(x, y int, erase bool) {
	c := d.Brush
	if erase {
		c = Transparent
	}
	d.stamp(x, y, c)
	if d.Symmetry {
		sx := d.Width - 1 - x
		if sx != x {
			d.stamp(sx, y, c)
		}
	}
}

func (d *Document) FillAt(x, y int) {
	d.flood(x, y, d.Brush)
	if d.Symmetry {
		d.flood(d.Width-1-x, y, d.Brush)
	}
}

func (d *Document) flood(x, y int, repl color.RGBA) {
	if !d.In(x, y) {
		return
	}
	target := d.At(x, y)
	if target == repl {
		return
	}
	w, h := d.Width, d.Height
	seen := make([]uint8, w*h)
	q := make([]int, 0, 64)
	start := y*w + x
	q = append(q, start)
	seen[start] = 1
	head := 0
	for head < len(q) {
		i := q[head]
		head++
		cx, cy := i%w, i/w
		if d.At(cx, cy) != target {
			continue
		}
		d.Img.SetRGBA(cx, cy, repl)
		try := func(nx, ny int) {
			if nx < 0 || ny < 0 || nx >= w || ny >= h {
				return
			}
			j := ny*w + nx
			if seen[j] == 1 {
				return
			}
			seen[j] = 1
			q = append(q, j)
		}
		try(cx+1, cy)
		try(cx-1, cy)
		try(cx, cy+1)
		try(cx, cy-1)
	}
}

func (d *Document) Pick(x, y int) bool {
	if !d.In(x, y) {
		return false
	}
	c := d.At(x, y)
	c.A = 255
	d.Brush = c
	d.Tool = ToolPen
	return true
}

func (d *Document) Apply(x, y int) {
	if !d.In(x, y) {
		return
	}
	switch d.Tool {
	case ToolPen:
		d.DrawBrush(x, y, false)
	case ToolEraser:
		d.DrawBrush(x, y, true)
	case ToolFill:
		d.FillAt(x, y)
	case ToolPipette:
		d.Pick(x, y)
	}
}

func (d *Document) MutatingTool() bool {
	return d.Tool == ToolPen || d.Tool == ToolEraser || d.Tool == ToolFill
}
