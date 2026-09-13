package capture

import (
	"image"
	"image/color"
	"testing"
)

func TestLimitedBufferRejectsOversizedOutput(t *testing.T) {
	buffer := limitedBuffer{remaining: 3}
	if _, err := buffer.Write([]byte("four")); err == nil {
		t.Fatal("expected oversized output to be rejected")
	}
	if buffer.Len() != 0 {
		t.Fatalf("buffer contains %d bytes after rejected write", buffer.Len())
	}
}

func TestSafeDimensions(t *testing.T) {
	tests := []struct {
		width, height int
		want          bool
	}{
		{1920, 1080, true},
		{15360, 6480, true},
		{10_000, 10_001, false},
		{0, 1080, false},
		{-1, 1080, false},
	}
	for _, tc := range tests {
		if got := safeDimensions(tc.width, tc.height); got != tc.want {
			t.Errorf("safeDimensions(%d, %d)=%v, want %v", tc.width, tc.height, got, tc.want)
		}
	}
}

func TestNewRejectsRelativeCustomCommand(t *testing.T) {
	if _, err := New([]string{"capture-helper"}); err == nil {
		t.Fatal("expected relative custom command to be rejected")
	}
}

func TestReleaseClearsPixelStorage(t *testing.T) {
	tests := []struct {
		name string
		img  image.Image
		data func() []byte
	}{
		{
			name: "RGBA",
			img:  image.NewRGBA(image.Rect(0, 0, 4, 4)),
			data: func() []byte { return nil },
		},
		{
			name: "NRGBA",
			img:  image.NewNRGBA(image.Rect(0, 0, 4, 4)),
			data: func() []byte { return nil },
		},
		{
			name: "YCbCr",
			img:  image.NewYCbCr(image.Rect(0, 0, 4, 4), image.YCbCrSubsampleRatio420),
			data: func() []byte { return nil },
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			switch img := tc.img.(type) {
			case *image.RGBA:
				for i := range img.Pix {
					img.Pix[i] = 0xff
				}
				tc.data = func() []byte { return img.Pix }
			case *image.NRGBA:
				for i := range img.Pix {
					img.Pix[i] = 0xff
				}
				tc.data = func() []byte { return img.Pix }
			case *image.YCbCr:
				for i := range img.Y {
					img.Y[i] = 0xff
				}
				for i := range img.Cb {
					img.Cb[i] = 0xff
				}
				for i := range img.Cr {
					img.Cr[i] = 0xff
				}
				tc.data = func() []byte { return append(append(img.Y, img.Cb...), img.Cr...) }
			}
			Release(tc.img)
			for i, value := range tc.data() {
				if value != 0 {
					t.Fatalf("pixel byte %d was not cleared", i)
				}
			}
		})
	}

	paletted := image.NewPaletted(image.Rect(0, 0, 2, 2), color.Palette{color.RGBA{1, 2, 3, 255}})
	for i := range paletted.Pix {
		paletted.Pix[i] = 1
	}
	Release(paletted)
	for _, value := range paletted.Pix {
		if value != 0 {
			t.Fatal("paletted pixel indexes were not cleared")
		}
	}
	if paletted.Palette[0] != nil {
		t.Fatal("palette was not cleared")
	}
}
