package capture

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os/exec"
)

type Capturer struct{ command []string }

func New(custom []string) (*Capturer, error) {
	if len(custom) > 0 {
		if _, err := exec.LookPath(custom[0]); err != nil {
			return nil, fmt.Errorf("capture command: %w", err)
		}
		return &Capturer{custom}, nil
	}
	candidates := [][]string{{"grim", "-"}, {"maim"}, {"scrot", "-"}}
	for _, c := range candidates {
		if _, err := exec.LookPath(c[0]); err == nil {
			return &Capturer{c}, nil
		}
	}
	return nil, fmt.Errorf("no screenshot tool found; install grim (Wayland), maim, or scrot (X11)")
}

func (c *Capturer) Capture(ctx context.Context) (image.Image, error) {
	cmd := exec.CommandContext(ctx, c.command[0], c.command[1:]...)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s failed: %w: %s", c.command[0], err, stderr.String())
	}
	img, _, err := image.Decode(bytes.NewReader(out.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("decode screenshot: %w", err)
	}
	return img, nil
}

func (c *Capturer) Name() string { return c.command[0] }
