package doc

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"path/filepath"
	"testing"
)

func TestNewIsTransparent(t *testing.T) {
	d := New(4, 4)
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			if d.At(x, y) != Transparent {
				t.Fatalf("pixel %d,%d = %#v, want transparent", x, y, d.At(x, y))
			}
		}
	}
}

func TestDrawBrushAndEraser(t *testing.T) {
	d := New(8, 8)
	d.Brush = color.RGBA{255, 0, 0, 255}
	d.SaveState()
	d.DrawBrush(2, 3, false)
	if d.At(2, 3) != d.Brush {
		t.Fatalf("pen did not paint, got %#v", d.At(2, 3))
	}
	d.DrawBrush(2, 3, true)
	if d.At(2, 3) != Transparent {
		t.Fatalf("eraser did not clear, got %#v", d.At(2, 3))
	}
}

func TestSymmetry(t *testing.T) {
	d := New(8, 8)
	d.Symmetry = true
	d.Brush = color.RGBA{0, 255, 0, 255}
	d.DrawBrush(1, 2, false)
	if d.At(1, 2) != d.Brush {
		t.Fatal("left pixel missing")
	}
	if d.At(6, 2) != d.Brush {
		t.Fatalf("mirrored pixel missing, got %#v", d.At(6, 2))
	}
}

func TestBrushSize(t *testing.T) {
	d := New(8, 8)
	d.Brush = color.RGBA{0, 0, 255, 255}
	d.BrushSize = 3
	d.DrawBrush(3, 3, false)
	painted := 0
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if d.At(x, y) == d.Brush {
				painted++
			}
		}
	}
	if painted != 9 {
		t.Fatalf("brush size 3 should paint 3x3=9 pixels, got %d", painted)
	}
}

func TestFloodFill(t *testing.T) {
	d := New(4, 4)
	d.Brush = color.RGBA{255, 0, 0, 255}
	d.FillAt(0, 0)
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			if d.At(x, y) != d.Brush {
				t.Fatalf("fill missed %d,%d", x, y)
			}
		}
	}
	d.FillAt(1, 1)
	if d.At(1, 1) != d.Brush {
		t.Fatal("same-color fill should be a no-op, not clear")
	}
}

func TestFloodFillRegion(t *testing.T) {
	d := New(4, 4)
	wall := color.RGBA{1, 2, 3, 255}
	for y := 0; y < 4; y++ {
		d.set(2, y, wall)
	}
	d.Brush = color.RGBA{9, 9, 9, 255}
	d.FillAt(0, 0)
	if d.At(0, 1) != d.Brush {
		t.Fatal("left region should fill")
	}
	if d.At(3, 1) != Transparent {
		t.Fatal("right region should stay empty")
	}
	if d.At(2, 1) != wall {
		t.Fatal("wall should be unchanged")
	}
}

func TestFloodFillSymmetry(t *testing.T) {
	d := New(6, 3)
	d.Symmetry = true
	d.Brush = color.RGBA{10, 20, 30, 255}
	d.set(1, 1, color.RGBA{1, 1, 1, 255})
	d.set(4, 1, color.RGBA{2, 2, 2, 255})
	d.FillAt(1, 1)
	if d.At(1, 1) != d.Brush {
		t.Fatal("seed not filled")
	}
	if d.At(4, 1) != d.Brush {
		t.Fatal("symmetric seed not filled")
	}
}

func TestPalette(t *testing.T) {
	d := New(3, 1)
	red := color.RGBA{255, 0, 0, 255}
	blue := color.RGBA{0, 0, 255, 255}
	d.set(0, 0, red)
	d.set(1, 0, red)
	d.set(2, 0, blue)
	pal := d.Palette()
	if len(pal) != 2 {
		t.Fatalf("palette len=%d, want 2", len(pal))
	}
	if pal[0] != red {
		t.Fatalf("most frequent should be red, got %#v", pal[0])
	}
	if pal[1] != blue {
		t.Fatalf("second should be blue, got %#v", pal[1])
	}
}

func TestUndoRedo(t *testing.T) {
	d := New(2, 2)
	d.Brush = color.RGBA{255, 255, 0, 255}
	d.SaveState()
	d.DrawBrush(0, 0, false)
	if d.At(0, 0) != d.Brush {
		t.Fatal("draw failed")
	}
	if !d.Undo() {
		t.Fatal("undo should succeed")
	}
	if d.At(0, 0) != Transparent {
		t.Fatal("undo should restore transparent pixel")
	}
	if !d.Redo() {
		t.Fatal("redo should succeed")
	}
	if d.At(0, 0) != d.Brush {
		t.Fatal("redo should restore yellow pixel")
	}
}

func TestZoomClamp(t *testing.T) {
	d := New(8, 8)
	d.SetZoom(0)
	if d.Zoom != MinZoom {
		t.Fatalf("zoom=%d", d.Zoom)
	}
	d.SetZoom(100)
	if d.Zoom != MaxZoom {
		t.Fatalf("zoom=%d", d.Zoom)
	}
	d.Zoom = 2
	d.ChangeZoom(-1)
	if d.Zoom != 1 {
		t.Fatalf("zoom=%d", d.Zoom)
	}
}

func TestPickSetsOpaquePen(t *testing.T) {
	d := New(2, 2)
	d.Tool = ToolPipette
	d.set(0, 0, color.RGBA{8, 16, 24, 0})
	if !d.Pick(0, 0) {
		t.Fatal("pick failed")
	}
	if d.Brush != (color.RGBA{8, 16, 24, 255}) {
		t.Fatalf("brush=%#v", d.Brush)
	}
	if d.Tool != ToolPen {
		t.Fatalf("tool=%s", d.Tool)
	}
}

func TestPNGRoundTrip(t *testing.T) {
	d := New(3, 2)
	d.Brush = color.RGBA{12, 34, 56, 255}
	d.DrawBrush(1, 1, false)

	dir := t.TempDir()
	path := filepath.Join(dir, "t.png")
	if err := d.Save(path); err != nil {
		t.Fatal(err)
	}
	img, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	got := ToRGBA(img)
	if got.RGBAAt(1, 1) != d.Brush {
		t.Fatalf("roundtrip pixel %#v", got.RGBAAt(1, 1))
	}
	if got.Bounds().Dx() != 3 || got.Bounds().Dy() != 2 {
		t.Fatalf("size %v", got.Bounds())
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, d.Img); err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(&buf)
	if err != nil {
		t.Fatal(err)
	}
	_ = decoded
}

func TestLoadImageSetsZoomForWideCanvas(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 120, 10))
	d := New(8, 8)
	d.Zoom = 10
	d.LoadImage(src, "wide.png")
	if d.Width != 120 || d.Height != 10 {
		t.Fatalf("size %dx%d", d.Width, d.Height)
	}
	if d.Zoom != 4 {
		t.Fatalf("zoom=%d, want 4", d.Zoom)
	}
	if d.FilePath != "wide.png" {
		t.Fatalf("path=%q", d.FilePath)
	}
}
