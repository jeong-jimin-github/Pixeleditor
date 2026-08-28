package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type swatch struct {
	widget.BaseWidget
	fill   color.Color
	min    fyne.Size
	onTap  func()
}

func newSwatch(c color.Color, min fyne.Size, onTap func()) *swatch {
	s := &swatch{fill: c, min: min, onTap: onTap}
	s.ExtendBaseWidget(s)
	return s
}

func (s *swatch) CreateRenderer() fyne.WidgetRenderer {
	r := canvas.NewRectangle(s.fill)
	r.CornerRadius = 3
	return widget.NewSimpleRenderer(r)
}

func (s *swatch) MinSize() fyne.Size {
	if s.min.Width > 0 {
		return s.min
	}
	return fyne.NewSize(28, 22)
}

func (s *swatch) Tapped(_ *fyne.PointEvent) {
	if s.onTap != nil {
		s.onTap()
	}
}

func (s *swatch) SetFill(c color.Color) {
	s.fill = c
	s.Refresh()
}
