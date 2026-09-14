package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	p := filepath.Join(t.TempDir(), "config.json")
	if err := WriteExample(p); err != nil {
		t.Fatal(err)
	}
	c, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if c.MinimumLengthPX != 40 || c.PollInterval.String() != "100ms" || c.CaptureFrameRate != 10 || c.DisappearanceConfirmationFrames != 3 || c.RestoreWithin.String() != "5m0s" || c.Color != (RGB{R: 255, G: 204, B: 0}) {
		t.Fatalf("unexpected config: %+v", c)
	}
}
func TestRejectsUnknownField(t *testing.T) {
	p := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(p, []byte(`{"unknown":true}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(p); err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadRejectsUnsafeFilesAndContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		mode    os.FileMode
	}{
		{name: "trailing JSON", content: `{}` + "\n{}", mode: 0600},
		{name: "group writable", content: `{}`, mode: 0620},
		{name: "relative capture command", content: `{"capture_command":["helper"]}`, mode: 0600},
		{name: "excessive color tolerance", content: `{"color_tolerance":65}`, mode: 0600},
		{name: "zero capture frame rate", content: `{"capture_frame_rate":0}`, mode: 0600},
		{name: "excessive capture frame rate", content: `{"capture_frame_rate":31}`, mode: 0600},
		{name: "zero disappearance confirmation", content: `{"disappearance_confirmation_frames":0}`, mode: 0600},
		{name: "excessive disappearance confirmation", content: `{"disappearance_confirmation_frames":31}`, mode: 0600},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(path, []byte(tc.content), tc.mode); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(path, tc.mode); err != nil {
				t.Fatal(err)
			}
			if _, err := Load(path); err == nil {
				t.Fatal("expected unsafe configuration to be rejected")
			}
		})
	}
}
