package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxConfigBytes = 1 << 20

type RGB struct{ R, G, B uint8 }

func (c *RGB) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return errors.New("color must be a string in #RRGGBB format")
	}
	if len(s) != 7 || s[0] != '#' {
		return errors.New("color must use #RRGGBB format")
	}
	var r, g, b uint8
	if _, err := fmt.Sscanf(s, "#%02x%02x%02x", &r, &g, &b); err != nil {
		return errors.New("invalid color")
	}
	*c = RGB{r, g, b}
	return nil
}

func (c RGB) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B))
}

type Config struct {
	Browsers                        []string `json:"browsers"`
	Color                           RGB      `json:"color"`
	ColorTolerance                  uint8    `json:"color_tolerance"`
	MinimumLengthPX                 int      `json:"minimum_length_px"`
	MinimumThicknessPX              int      `json:"minimum_thickness_px"`
	PollInterval                    Duration `json:"poll_interval"`
	CaptureFrameRate                int      `json:"capture_frame_rate"`
	DisappearanceConfirmationFrames int      `json:"disappearance_confirmation_frames"`
	RestoreWithin                   Duration `json:"restore_within"`
	CaptureCommand                  []string `json:"capture_command,omitempty"`
	AudioBackend                    string   `json:"audio_backend"`
	LogLevel                        string   `json:"log_level"`
}

type Duration struct{ time.Duration }

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return errors.New("duration must be a string")
	}
	v, err := time.ParseDuration(s)
	d.Duration = v
	return err
}
func (d Duration) MarshalJSON() ([]byte, error) { return json.Marshal(d.String()) }

func Defaults() Config {
	return Config{
		Browsers: []string{"firefox", "google-chrome", "chromium", "brave-browser"},
		Color:    RGB{255, 204, 0}, ColorTolerance: 24, MinimumLengthPX: 40, MinimumThicknessPX: 2,
		PollInterval: Duration{100 * time.Millisecond}, CaptureFrameRate: 10,
		DisappearanceConfirmationFrames: 3, RestoreWithin: Duration{5 * time.Minute},
		AudioBackend: "auto", LogLevel: "info",
	}
}

func DefaultPath() string {
	if p := os.Getenv("XDG_CONFIG_HOME"); p != "" {
		return filepath.Join(p, "yellow-silence", "config.json")
	}
	h, err := os.UserHomeDir()
	if err != nil {
		return "config.json"
	}
	return filepath.Join(h, ".config", "yellow-silence", "config.json")
}

func Load(path string) (Config, error) {
	cfg := Defaults()
	file, err := os.Open(path)
	if err != nil {
		return cfg, fmt.Errorf("open config: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return cfg, fmt.Errorf("inspect config: %w", err)
	}
	if !info.Mode().IsRegular() {
		return cfg, errors.New("config must be a regular file")
	}
	if info.Size() > maxConfigBytes {
		return cfg, errors.New("config exceeds 1 MiB safety limit")
	}
	if info.Mode().Perm()&0022 != 0 {
		return cfg, errors.New("config must not be writable by group or other users")
	}
	b, err := io.ReadAll(io.LimitReader(file, maxConfigBytes+1))
	if err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if len(b) > maxConfigBytes {
		return cfg, errors.New("config exceeds 1 MiB safety limit")
	}
	dec := json.NewDecoder(strings.NewReader(string(b)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return cfg, errors.New("config must contain exactly one JSON object")
	}
	if err := cfg.Validate(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if len(c.Browsers) == 0 || len(c.Browsers) > 64 {
		return errors.New("browsers must contain between 1 and 64 entries")
	}
	for _, b := range c.Browsers {
		if strings.TrimSpace(b) == "" || len(b) > 4096 || strings.ContainsAny(b, "\x00\r\n") {
			return errors.New("browser entries must be non-empty process names or executable paths")
		}
	}
	if c.ColorTolerance > 64 {
		return errors.New("color_tolerance must not exceed 64")
	}
	if c.MinimumLengthPX < 1 || c.MinimumThicknessPX < 1 {
		return errors.New("minimum dimensions must be positive")
	}
	if c.PollInterval.Duration < 100*time.Millisecond {
		return errors.New("poll_interval must be at least 100ms")
	}
	if c.CaptureFrameRate < 1 || c.CaptureFrameRate > 30 {
		return errors.New("capture_frame_rate must be between 1 and 30")
	}
	if c.DisappearanceConfirmationFrames < 1 || c.DisappearanceConfirmationFrames > 30 {
		return errors.New("disappearance_confirmation_frames must be between 1 and 30")
	}
	if c.RestoreWithin.Duration < 0 {
		return errors.New("restore_within cannot be negative")
	}
	if c.AudioBackend != "auto" && c.AudioBackend != "pactl" && c.AudioBackend != "wpctl" {
		return errors.New("audio_backend must be auto, pactl, or wpctl")
	}
	if c.LogLevel != "debug" && c.LogLevel != "info" && c.LogLevel != "warn" && c.LogLevel != "error" {
		return errors.New("invalid log_level")
	}
	if len(c.CaptureCommand) > 64 {
		return errors.New("capture_command must not contain more than 64 arguments")
	}
	if len(c.CaptureCommand) > 0 {
		if !filepath.IsAbs(c.CaptureCommand[0]) {
			return errors.New("capture_command executable must be an absolute path")
		}
		for _, arg := range c.CaptureCommand {
			if len(arg) > 4096 || strings.ContainsRune(arg, '\x00') {
				return errors.New("capture_command arguments must not exceed 4096 bytes or contain NUL")
			}
		}
	}
	return nil
}

func WriteExample(path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("refusing to overwrite existing %s", path)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(Defaults(), "", "  ")
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0600); err != nil {
		return err
	}
	fmt.Printf("created %s\n", path)
	return nil
}
