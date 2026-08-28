package doc

import (
	"image/color"
	"sort"
)

func colorKey(c color.RGBA) uint32 {
	return uint32(c.R)<<24 | uint32(c.G)<<16 | uint32(c.B)<<8 | uint32(c.A)
}

// Palette returns unique opaque colors in the image, most frequent first.
func (d *Document) Palette() []color.RGBA {
	counts := make(map[uint32]int)
	colors := make(map[uint32]color.RGBA)
	for y := 0; y < d.Height; y++ {
		for x := 0; x < d.Width; x++ {
			c := d.At(x, y)
			if c.A == 0 {
				continue
			}
			k := colorKey(c)
			counts[k]++
			colors[k] = c
		}
	}
	type pair struct {
		key   uint32
		count int
	}
	list := make([]pair, 0, len(counts))
	for k, n := range counts {
		list = append(list, pair{k, n})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].count == list[j].count {
			return list[i].key < list[j].key
		}
		return list[i].count > list[j].count
	})
	out := make([]color.RGBA, len(list))
	for i, p := range list {
		out[i] = colors[p.key]
	}
	return out
}
