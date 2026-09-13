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
	if c.MinimumLengthPX != 100 || c.RestoreWithin.String() != "5m0s" || c.Color != (RGB{R: 255, G: 204, B: 0}) {
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
