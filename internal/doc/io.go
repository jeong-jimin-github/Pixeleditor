package doc

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/bmp"
)

func Load(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Decode(f)
}

func Decode(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	return img, err
}

func (d *Document) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := Encode(f, d.Img, path); err != nil {
		return err
	}
	d.FilePath = path
	return nil
}

func Encode(w io.Writer, img image.Image, path string) error {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg":
		return jpeg.Encode(w, img, &jpeg.Options{Quality: 95})
	case ".bmp":
		return bmp.Encode(w, img)
	case ".png", "":
		return png.Encode(w, img)
	default:
		return fmt.Errorf("unsupported format: %s", filepath.Ext(path))
	}
}

func init() {
	image.RegisterFormat("bmp", "BM", bmp.Decode, bmp.DecodeConfig)
}
