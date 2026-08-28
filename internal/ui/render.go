package ui

import (
	"image"
	"image/color"

	"github.com/jeong-jimin-github/Pixeleditor/internal/doc"
)

var (
	checkerLight = color.RGBA{255, 255, 255, 255}
	checkerDark  = color.RGBA{204, 204, 204, 255}
	gridColor    = color.RGBA{0x88, 0x88, 0x88, 255}
	symColor     = color.RGBA{0, 90, 220, 255}
	cursorColor  = color.RGBA{220, 40, 40, 255}
)

func over(src, dst color.RGBA) color.RGBA {
	if src.A == 255 {
		return src
	}
	if src.A == 0 {
		return dst
	}
	a := uint32(src.A)
	ia := 255 - a
	return color.RGBA{
		R: uint8((uint32(src.R)*a + uint32(dst.R)*ia) / 255),
		G: uint8((uint32(src.G)*a + uint32(dst.G)*ia) / 255),
		B: uint8((uint32(src.B)*a + uint32(dst.B)*ia) / 255),
		A: 255,
	}
}

func composite(d *doc.Document) *image.RGBA {
	out := image.NewRGBA(image.Rect(0, 0, d.Width, d.Height))
	for y := 0; y < d.Height; y++ {
		for x := 0; x < d.Width; x++ {
			bg := checkerLight
			if (x+y)%2 == 1 {
				bg = checkerDark
			}
			out.SetRGBA(x, y, over(d.At(x, y), bg))
		}
	}
	return out
}

func scaleNearest(src *image.RGBA, z int) *image.RGBA {
	if z < 1 {
		z = 1
	}
	b := src.Bounds()
	w, h := b.Dx()*z, b.Dy()*z
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			c := src.RGBAAt(x, y)
			for oy := 0; oy < z; oy++ {
				row := dst.Pix[(y*z+oy)*dst.Stride+x*z*4:]
				for ox := 0; ox < z; ox++ {
					i := ox * 4
					row[i+0] = c.R
					row[i+1] = c.G
					row[i+2] = c.B
					row[i+3] = c.A
				}
			}
		}
	}
	return dst
}

func hline(img *image.RGBA, y, w int, c color.RGBA) {
	if y < 0 || y >= img.Bounds().Dy() {
		return
	}
	if w > img.Bounds().Dx() {
		w = img.Bounds().Dx()
	}
	for x := 0; x < w; x++ {
		img.SetRGBA(x, y, c)
	}
}

func vline(img *image.RGBA, x, h int, c color.RGBA) {
	if x < 0 || x >= img.Bounds().Dx() {
		return
	}
	if h > img.Bounds().Dy() {
		h = img.Bounds().Dy()
	}
	for y := 0; y < h; y++ {
		img.SetRGBA(x, y, c)
	}
}

func plot(img *image.RGBA, x, y int, c color.RGBA) {
	if x < 0 || y < 0 || x >= img.Bounds().Dx() || y >= img.Bounds().Dy() {
		return
	}
	img.SetRGBA(x, y, c)
}

func rectOutline(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA, dash bool) {
	if x1 < x0 {
		x0, x1 = x1, x0
	}
	if y1 < y0 {
		y0, y1 = y1, y0
	}
	on := func(n int) bool {
		if !dash {
			return true
		}
		return (n/3)%2 == 0
	}
	n := 0
	for x := x0; x <= x1; x++ {
		if on(n) {
			plot(img, x, y0, c)
			plot(img, x, y1, c)
		}
		n++
	}
	n = 0
	for y := y0; y <= y1; y++ {
		if on(n) {
			plot(img, x0, y, c)
			plot(img, x1, y, c)
		}
		n++
	}
}

func render(d *doc.Document, hoverX, hoverY int) *image.RGBA {
	z := d.Zoom
	if z < 1 {
		z = 1
	}
	out := scaleNearest(composite(d), z)
	w, h := d.Width*z, d.Height*z

	if d.ShowGrid {
		step := d.GridGap * z
		if step >= 4 {
			for x := 0; x <= w; x += step {
				vline(out, x, h, gridColor)
			}
			for y := 0; y <= h; y += step {
				hline(out, y, w, gridColor)
			}
		}
	}
	if d.Symmetry {
		cx := (d.Width / 2) * z
		vline(out, cx, h, symColor)
		if cx-1 >= 0 {
			vline(out, cx-1, h, symColor)
		}
	}

	if hoverX >= 0 && hoverY >= 0 && hoverX < d.Width && hoverY < d.Height {
		bs := d.BrushSize
		if bs < 1 {
			bs = 1
		}
		offset := (bs / 2) * z
		sx := hoverX*z - offset
		sy := hoverY*z - offset
		ex := sx + z*bs - 1
		ey := sy + z*bs - 1
		rectOutline(out, sx, sy, ex, ey, cursorColor, false)
		if d.Symmetry {
			symX := d.Width - 1 - hoverX
			sx2 := symX*z - offset
			rectOutline(out, sx2, sy, sx2+z*bs-1, ey, symColor, true)
		}
	}
	return out
}
