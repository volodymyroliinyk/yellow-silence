package audio

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const (
	audioCommandTimeout = 3 * time.Second
	maxAudioOutputBytes = 4096
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
		path, err := exec.LookPath(name)
		if err != nil {
			continue
		}
		switch name {
		case "wpctl":
			return &commandController{name: name, get: []string{path, "get-volume", "@DEFAULT_AUDIO_SINK@"}, setMute: []string{path, "set-mute", "@DEFAULT_AUDIO_SINK@", "1"}, setUnmute: []string{path, "set-mute", "@DEFAULT_AUDIO_SINK@", "0"}}, nil
		case "pactl":
			return &commandController{name: name, get: []string{path, "get-sink-mute", "@DEFAULT_SINK@"}, setMute: []string{path, "set-sink-mute", "@DEFAULT_SINK@", "1"}, setUnmute: []string{path, "set-sink-mute", "@DEFAULT_SINK@", "0"}}, nil
		}
	}
	return nil, fmt.Errorf("no audio controller found; install wpctl (recommended) or pactl")
}
func (c *commandController) Name() string { return c.name }
func (c *commandController) Muted(ctx context.Context) (bool, error) {
	b, err := commandOutput(ctx, c.get)
	if err != nil {
		return false, fmt.Errorf("query mute state: %w", err)
	}
	s := strings.ToLower(string(b))
	return strings.Contains(s, "muted") || strings.Contains(s, "mute: yes"), nil
}
func (c *commandController) SetMuted(ctx context.Context, m bool) error {
	a := c.setUnmute
	if m {
		a = c.setMute
	}
	_, err := commandOutput(ctx, a)
	if err != nil {
		return fmt.Errorf("set mute=%t: %w", m, err)
	}
	return nil
}

func commandOutput(ctx context.Context, args []string) ([]byte, error) {
	commandCtx, cancel := context.WithTimeout(ctx, audioCommandTimeout)
	defer cancel()
	cmd := exec.CommandContext(commandCtx, args[0], args[1:]...)
	out := boundedOutput{remaining: maxAudioOutputBytes}
	cmd.Stdout = &out
	cmd.Stderr = nil
	err := cmd.Run()
	if commandCtx.Err() != nil {
		return nil, commandCtx.Err()
	}
	return out.Bytes(), err
}

type boundedOutput struct {
	bytes.Buffer
	remaining int
}

func (b *boundedOutput) Write(p []byte) (int, error) {
	if len(p) > b.remaining {
		return 0, errors.New("command output exceeds safety limit")
	}
	n, err := b.Buffer.Write(p)
	b.remaining -= n
	return n, err
}
