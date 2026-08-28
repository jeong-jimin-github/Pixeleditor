package ui

import (
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"github.com/jeong-jimin-github/Pixeleditor/internal/doc"
)

type PixelView struct {
	widget.BaseWidget
	doc       *doc.Document
	scroll    *container.Scroll
	win       fyne.Window
	hoverX    int
	hoverY    int
	painting  bool
	panning   bool
	spaceDown bool
	middlePan bool
	onChange  func()
	onHover   func(x, y int)
}

func newPixelView(d *doc.Document) *PixelView {
	v := &PixelView{doc: d, hoverX: -1, hoverY: -1}
	v.ExtendBaseWidget(v)
	return v
}

func (v *PixelView) CreateRenderer() fyne.WidgetRenderer {
	img := canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 1, 1)))
	img.ScaleMode = canvas.ImageScalePixels
	img.FillMode = canvas.ImageFillOriginal
	img.SetMinSize(v.MinSize())
	return &pixelRenderer{view: v, img: img}
}

func (v *PixelView) MinSize() fyne.Size {
	z := float32(v.doc.Zoom)
	if z < 1 {
		z = 1
	}
	return fyne.NewSize(float32(v.doc.Width)*z, float32(v.doc.Height)*z)
}

func (v *PixelView) pixelAt(pos fyne.Position) (int, int) {
	z := float32(v.doc.Zoom)
	if z < 1 {
		z = 1
	}
	return int(pos.X / z), int(pos.Y / z)
}

func (v *PixelView) FocusGained() {}
func (v *PixelView) FocusLost() {
	v.spaceDown = false
	v.panning = false
}
func (v *PixelView) TypedRune(_ rune) {}
func (v *PixelView) TypedKey(_ *fyne.KeyEvent) {}

func (v *PixelView) KeyDown(ev *fyne.KeyEvent) {
	if ev.Name == fyne.KeySpace {
		v.spaceDown = true
		v.panning = true
	}
}

func (v *PixelView) KeyUp(ev *fyne.KeyEvent) {
	if ev.Name == fyne.KeySpace {
		v.spaceDown = false
		if !v.middlePan {
			v.panning = false
		}
	}
}

func (v *PixelView) MouseIn(ev *desktop.MouseEvent) {
	if v.win != nil {
		v.win.Canvas().Focus(v)
	}
	v.MouseMoved(ev)
}

func (v *PixelView) MouseMoved(ev *desktop.MouseEvent) {
	x, y := v.pixelAt(ev.Position)
	if x == v.hoverX && y == v.hoverY {
		return
	}
	v.hoverX, v.hoverY = x, y
	v.Refresh()
	if v.onHover != nil {
		v.onHover(x, y)
	}
}

func (v *PixelView) MouseOut() {
	v.hoverX, v.hoverY = -1, -1
	v.Refresh()
	if v.onHover != nil {
		v.onHover(-1, -1)
	}
}

func (v *PixelView) MouseDown(ev *desktop.MouseEvent) {
	if v.win != nil {
		v.win.Canvas().Focus(v)
	}
	switch ev.Button {
	case desktop.MouseButtonSecondary:
		x, y := v.pixelAt(ev.Position)
		if v.doc.Pick(x, y) && v.onChange != nil {
			v.onChange()
		}
		v.Refresh()
		return
	case desktop.MouseButtonTertiary:
		v.middlePan = true
		v.panning = true
		return
	case desktop.MouseButtonPrimary:
		if v.panning || v.spaceDown {
			return
		}
		x, y := v.pixelAt(ev.Position)
		if v.doc.MutatingTool() {
			v.doc.SaveState()
			v.painting = true
			v.doc.Apply(x, y)
			v.Refresh()
			if v.onChange != nil {
				v.onChange()
			}
			return
		}
		if v.doc.Tool == doc.ToolPipette {
			if v.doc.Pick(x, y) && v.onChange != nil {
				v.onChange()
			}
			v.Refresh()
		}
	}
}

func (v *PixelView) MouseUp(ev *desktop.MouseEvent) {
	if ev.Button == desktop.MouseButtonTertiary {
		v.middlePan = false
		if !v.spaceDown {
			v.panning = false
		}
	}
	if ev.Button == desktop.MouseButtonPrimary {
		v.painting = false
	}
}

func (v *PixelView) Dragged(ev *fyne.DragEvent) {
	if v.panning && v.scroll != nil {
		off := v.scroll.Offset
		v.scroll.Offset = fyne.NewPos(off.X-ev.Dragged.DX, off.Y-ev.Dragged.DY)
		v.scroll.Refresh()
		return
	}
	if !v.painting {
		return
	}
	x, y := v.pixelAt(ev.Position)
	v.hoverX, v.hoverY = x, y
	v.doc.Apply(x, y)
	v.Refresh()
	if v.onHover != nil {
		v.onHover(x, y)
	}
}

func (v *PixelView) DragEnd() {
	v.painting = false
}

func (v *PixelView) Scrolled(ev *fyne.ScrollEvent) {
	if ev.Scrolled.DY > 0 {
		v.doc.ChangeZoom(1)
	} else if ev.Scrolled.DY < 0 {
		v.doc.ChangeZoom(-1)
	}
	v.Refresh()
	if v.onChange != nil {
		v.onChange()
	}
}

type pixelRenderer struct {
	view *PixelView
	img  *canvas.Image
}

func (r *pixelRenderer) Destroy() {}

func (r *pixelRenderer) Layout(s fyne.Size) {
	r.img.Resize(s)
}

func (r *pixelRenderer) MinSize() fyne.Size {
	return r.view.MinSize()
}

func (r *pixelRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.img}
}

func (r *pixelRenderer) Refresh() {
	rendered := render(r.view.doc, r.view.hoverX, r.view.hoverY)
	r.img.Image = rendered
	r.img.SetMinSize(r.view.MinSize())
	r.img.Refresh()
}

var (
	_ fyne.Widget         = (*PixelView)(nil)
	_ fyne.Draggable      = (*PixelView)(nil)
	_ fyne.Focusable      = (*PixelView)(nil)
	_ fyne.Scrollable     = (*PixelView)(nil)
	_ desktop.Hoverable   = (*PixelView)(nil)
	_ desktop.Mouseable   = (*PixelView)(nil)
	_ desktop.Keyable     = (*PixelView)(nil)
)
