package detect

import (
	"github.com/volodymyr/yellow-silence/internal/config"
	"image"
)

type Match struct{ X, Y, Length, Thickness int }

func FindHorizontalBar(img image.Image, target config.RGB, tolerance uint8, minLength, minThickness int) (Match, bool) {
	b := img.Bounds()
	if minLength < 1 || minThickness < 1 || b.Dx() < minLength || b.Dy() < minThickness {
		return Match{}, false
	}
	prev := make([]int, b.Dx())
	for y := b.Min.Y; y < b.Max.Y; y++ {
		run := 0
		for x := b.Min.X; x < b.Max.X; x++ {
			if closeColor(img.At(x, y), target, tolerance) {
				run++
			} else {
				run = 0
			}
			i := x - b.Min.X
			if run >= minLength {
				prev[i]++
			} else {
				prev[i] = 0
			}
			if run >= minLength && prev[i] >= minThickness {
				return Match{X: x - run + 1, Y: y - prev[i] + 1, Length: run, Thickness: prev[i]}, true
			}
		}
	}
	return Match{}, false
}

func closeColor(c interface {
	RGBA() (uint32, uint32, uint32, uint32)
}, t config.RGB, tol uint8) bool {
	r, g, b, a := c.RGBA()
	if a < 0x8000 {
		return false
	}
	return delta(uint8(r>>8), t.R) <= tol && delta(uint8(g>>8), t.G) <= tol && delta(uint8(b>>8), t.B) <= tol
}
func delta(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}
