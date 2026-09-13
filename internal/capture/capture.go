package capture

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	maxCaptureBytes  = 64 << 20
	maxCapturePixels = 100_000_000
	captureTimeout   = 5 * time.Second
)

type Capturer struct{ command []string }

func New(custom []string) (*Capturer, error) {
	if len(custom) > 0 {
		if !filepath.IsAbs(custom[0]) {
			return nil, fmt.Errorf("custom capture command must use an absolute executable path")
		}
		path, err := exec.LookPath(custom[0])
		if err != nil {
			return nil, fmt.Errorf("capture command: %w", err)
		}
		command := append([]string{path}, custom[1:]...)
		return &Capturer{command}, nil
	}
	candidates := [][]string{{"grim", "-"}, {"maim"}, {"scrot", "-"}}
	for _, c := range candidates {
		if path, err := exec.LookPath(c[0]); err == nil {
			command := append([]string{path}, c[1:]...)
			return &Capturer{command}, nil
		}
	}
	return nil, fmt.Errorf("no screenshot tool found; install grim (Wayland), maim, or scrot (X11)")
}

func (c *Capturer) Capture(ctx context.Context) (image.Image, error) {
	commandCtx, cancel := context.WithTimeout(ctx, captureTimeout)
	defer cancel()
	cmd := exec.CommandContext(commandCtx, c.command[0], c.command[1:]...)
	var out limitedBuffer
	out.remaining = maxCaptureBytes
	cmd.Stdout = &out
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		if commandCtx.Err() != nil {
			return nil, fmt.Errorf("capture command timed out or was canceled: %w", commandCtx.Err())
		}
		return nil, fmt.Errorf("capture command failed: %w", err)
	}
	encoded := out.Bytes()
	defer clear(encoded)
	metadata, _, err := image.DecodeConfig(bytes.NewReader(encoded))
	if err != nil {
		return nil, fmt.Errorf("decode screenshot metadata: %w", err)
	}
	if !safeDimensions(metadata.Width, metadata.Height) {
		return nil, fmt.Errorf("screenshot dimensions exceed safety limit")
	}
	img, _, err := image.Decode(bytes.NewReader(encoded))
	if err != nil {
		return nil, fmt.Errorf("decode screenshot: %w", err)
	}
	return img, nil
}

func safeDimensions(width, height int) bool {
	return width > 0 && height > 0 && uint64(width)*uint64(height) <= maxCapturePixels
}

// Release clears pixel storage owned by decoded image types used by the
// standard PNG and JPEG decoders. Call it as soon as detection is complete.
func Release(img image.Image) {
	switch pixels := img.(type) {
	case *image.RGBA:
		clear(pixels.Pix)
	case *image.RGBA64:
		clear(pixels.Pix)
	case *image.NRGBA:
		clear(pixels.Pix)
	case *image.NRGBA64:
		clear(pixels.Pix)
	case *image.Alpha:
		clear(pixels.Pix)
	case *image.Alpha16:
		clear(pixels.Pix)
	case *image.Gray:
		clear(pixels.Pix)
	case *image.Gray16:
		clear(pixels.Pix)
	case *image.CMYK:
		clear(pixels.Pix)
	case *image.Paletted:
		clear(pixels.Pix)
		clear(pixels.Palette)
	case *image.YCbCr:
		clear(pixels.Y)
		clear(pixels.Cb)
		clear(pixels.Cr)
	}
}

func (c *Capturer) Name() string { return c.command[0] }

type limitedBuffer struct {
	bytes.Buffer
	remaining int
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if len(p) > b.remaining {
		return 0, fmt.Errorf("capture output exceeds %d bytes", maxCaptureBytes)
	}
	n, err := b.Buffer.Write(p)
	b.remaining -= n
	return n, err
}
