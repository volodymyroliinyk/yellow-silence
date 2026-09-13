package audio

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type Controller interface {
	Muted(context.Context) (bool, error)
	SetMuted(context.Context, bool) error
	Name() string
}
type commandController struct {
	name                    string
	get, setMute, setUnmute []string
}

func New(preference string) (Controller, error) {
	order := []string{"wpctl", "pactl"}
	if preference != "auto" {
		order = []string{preference}
	}
	for _, name := range order {
		if _, err := exec.LookPath(name); err != nil {
			continue
		}
		switch name {
		case "wpctl":
			return &commandController{name: name, get: []string{"wpctl", "get-volume", "@DEFAULT_AUDIO_SINK@"}, setMute: []string{"wpctl", "set-mute", "@DEFAULT_AUDIO_SINK@", "1"}, setUnmute: []string{"wpctl", "set-mute", "@DEFAULT_AUDIO_SINK@", "0"}}, nil
		case "pactl":
			return &commandController{name: name, get: []string{"pactl", "get-sink-mute", "@DEFAULT_SINK@"}, setMute: []string{"pactl", "set-sink-mute", "@DEFAULT_SINK@", "1"}, setUnmute: []string{"pactl", "set-sink-mute", "@DEFAULT_SINK@", "0"}}, nil
		}
	}
	return nil, fmt.Errorf("no audio controller found; install wpctl (recommended) or pactl")
}
func (c *commandController) Name() string { return c.name }
func (c *commandController) Muted(ctx context.Context) (bool, error) {
	b, err := exec.CommandContext(ctx, c.get[0], c.get[1:]...).CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("query mute state: %w: %s", err, b)
	}
	s := strings.ToLower(string(b))
	return strings.Contains(s, "muted") || strings.Contains(s, "mute: yes"), nil
}
func (c *commandController) SetMuted(ctx context.Context, m bool) error {
	a := c.setUnmute
	if m {
		a = c.setMute
	}
	b, err := exec.CommandContext(ctx, a[0], a[1:]...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("set mute=%t: %w: %s", m, err, b)
	}
	return nil
}
