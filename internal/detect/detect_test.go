package detect

import (
	"github.com/volodymyr/yellow-silence/internal/config"
	"image"
	"image/color"
	"testing"
)

func TestFindHorizontalBar(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 30))
	for y := 10; y < 13; y++ {
		for x := 20; x < 141; x++ {
			img.Set(x, y, color.RGBA{250, 210, 5, 255})
		}
	}
	m, ok := FindHorizontalBar(img, config.RGB{R: 255, G: 215, B: 0}, 10, 100, 2)
	if !ok || m.Length < 100 || m.Thickness < 2 {
		t.Fatalf("match=%+v ok=%v", m, ok)
	}
}

func TestFindVideoProgressBarExamples(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{name: "full width", length: 1340},
		{name: "partial progress", length: 1019},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, 1374, 587))
			for y := 488; y < 491; y++ {
				for x := 34; x < 34+tc.length; x++ {
					img.Set(x, y, color.RGBA{255, 204, 0, 255})
				}
			}

			match, ok := FindHorizontalBar(img, config.RGB{R: 255, G: 204, B: 0}, 24, 100, 2)
			if !ok {
				t.Fatal("expected video progress bar to match")
			}
			if match.X != 34 || match.Y != 488 || match.Length < 100 || match.Thickness != 2 {
				t.Fatalf("unexpected match: %+v", match)
			}
		})
	}
}
func TestRejectsShortAndWrongColor(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 20))
	for x := 0; x < 99; x++ {
		img.Set(x, 5, color.RGBA{255, 215, 0, 255})
	}
	if _, ok := FindHorizontalBar(img, config.RGB{R: 255, G: 215, B: 0}, 0, 100, 1); ok {
		t.Fatal("short bar matched")
	}
	for x := 0; x < 150; x++ {
		img.Set(x, 8, color.RGBA{255, 0, 0, 255})
	}
	if _, ok := FindHorizontalBar(img, config.RGB{R: 255, G: 215, B: 0}, 10, 100, 1); ok {
		t.Fatal("wrong color matched")
	}
}
