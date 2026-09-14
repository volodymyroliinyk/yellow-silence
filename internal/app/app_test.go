package app

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/volodymyroliinyk/yellow-silence/internal/config"
)

type fakeAudio struct {
	setCalls []bool
	setErr   error
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func (f *fakeAudio) Muted(context.Context) (bool, error) { return false, nil }
func (f *fakeAudio) Name() string                        { return "fake" }
func (f *fakeAudio) SetMuted(_ context.Context, muted bool) error {
	f.setCalls = append(f.setCalls, muted)
	return f.setErr
}

func TestRestoreNeverUnmutesWithoutOwnership(t *testing.T) {
	audio := &fakeAudio{}
	a := &App{cfg: config.Config{RestoreWithin: config.Duration{Duration: 5 * time.Minute}}, log: testLogger(), audio: audio}
	if err := a.restoreIfAllowed(context.Background(), true); err != nil {
		t.Fatal(err)
	}
	if len(audio.setCalls) != 0 {
		t.Fatalf("unexpected audio changes: %v", audio.setCalls)
	}
}

func TestRestoreRespectsDeadline(t *testing.T) {
	tests := []struct {
		name       string
		mutedAt    time.Time
		wantUnmute bool
	}{
		{name: "inside restore window", mutedAt: time.Now().Add(-time.Minute), wantUnmute: true},
		{name: "expired restore window", mutedAt: time.Now().Add(-6 * time.Minute), wantUnmute: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			audio := &fakeAudio{}
			a := &App{
				cfg:       config.Config{RestoreWithin: config.Duration{Duration: 5 * time.Minute}},
				log:       testLogger(),
				audio:     audio,
				mutedByUs: true,
				mutedAt:   tc.mutedAt,
			}
			if err := a.restoreIfAllowed(context.Background(), true); err != nil {
				t.Fatal(err)
			}
			gotUnmute := len(audio.setCalls) == 1 && !audio.setCalls[0]
			if gotUnmute != tc.wantUnmute {
				t.Fatalf("unmute=%v, want %v; calls=%v", gotUnmute, tc.wantUnmute, audio.setCalls)
			}
		})
	}
}

func TestFailedRestoreRetainsOwnershipForRetry(t *testing.T) {
	audio := &fakeAudio{setErr: errors.New("audio unavailable")}
	a := &App{
		cfg:       config.Config{RestoreWithin: config.Duration{Duration: 5 * time.Minute}},
		log:       testLogger(),
		audio:     audio,
		mutedByUs: true,
		mutedAt:   time.Now(),
	}
	if err := a.restoreIfAllowed(context.Background(), true); err == nil {
		t.Fatal("expected restore error")
	}
	if !a.mutedByUs {
		t.Fatal("mute ownership was lost after failed restore")
	}
}

func TestTransientMissingFramesDoNotRestoreAudio(t *testing.T) {
	audio := &fakeAudio{}
	a := &App{
		cfg:       config.Config{RestoreWithin: config.Duration{Duration: 5 * time.Minute}},
		log:       testLogger(),
		audio:     audio,
		mutedByUs: true,
		mutedAt:   time.Now(),
	}
	for frame := 1; frame <= disappearanceConfirmationFrames; frame++ {
		confirmed := a.disappearanceConfirmed()
		if confirmed != (frame == disappearanceConfirmationFrames) {
			t.Fatalf("frame %d: confirmed=%v", frame, confirmed)
		}
	}
	if len(audio.setCalls) != 0 || !a.mutedByUs {
		t.Fatalf("transient misses changed audio: calls=%v owned=%v", audio.setCalls, a.mutedByUs)
	}
	if err := a.restoreIfAllowed(context.Background(), true); err != nil {
		t.Fatal(err)
	}
	if len(audio.setCalls) != 1 || audio.setCalls[0] || a.mutedByUs {
		t.Fatalf("confirmed disappearance did not restore: calls=%v owned=%v", audio.setCalls, a.mutedByUs)
	}
}
