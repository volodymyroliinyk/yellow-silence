package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/volodymyroliinyk/yellow-silence/internal/audio"
	"github.com/volodymyroliinyk/yellow-silence/internal/capture"
	"github.com/volodymyroliinyk/yellow-silence/internal/config"
	"github.com/volodymyroliinyk/yellow-silence/internal/detect"
	proc "github.com/volodymyroliinyk/yellow-silence/internal/process"
)

type App struct {
	cfg           config.Config
	log           *slog.Logger
	audio         audio.Controller
	capture       *capture.Capturer
	mutedByUs     bool
	mutedAt       time.Time
	missingFrames int
}

const disappearanceConfirmationFrames = 3

func New(cfg config.Config, log *slog.Logger) *App { return &App{cfg: cfg, log: log} }

func (a *App) Run(ctx context.Context) error {
	var err error
	a.capture, err = capture.New(a.cfg.CaptureCommand)
	if err != nil {
		return err
	}
	defer a.capture.Close()
	a.audio, err = audio.New(a.cfg.AudioBackend)
	if err != nil {
		return err
	}
	a.log.Info("service started", "capture_backend", a.capture.Name(), "audio_backend", a.audio.Name(), "poll_interval", a.cfg.PollInterval.String())
	defer func() { a.log.Info("service stopped") }()
	if err := a.tick(ctx); err != nil {
		a.log.Warn("monitor cycle failed", "error", err)
	}
	t := time.NewTicker(a.cfg.PollInterval.Duration)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			if err := a.tick(ctx); err != nil {
				a.log.Warn("monitor cycle failed", "error", err)
			}
		}
	}
}
func (a *App) tick(ctx context.Context) error {
	if !proc.AnyRunning(a.cfg.Browsers) {
		a.log.Debug("no configured browser is running")
		a.missingFrames = 0
		return a.restoreIfAllowed(ctx, false)
	}
	img, err := a.capture.Capture(ctx)
	if err != nil {
		return err
	}
	m, found := func() (detect.Match, bool) {
		defer capture.Release(img)
		return detect.FindHorizontalBar(img, a.cfg.Color, a.cfg.ColorTolerance, a.cfg.MinimumLengthPX, a.cfg.MinimumThicknessPX)
	}()
	if found {
		a.missingFrames = 0
		a.log.Debug("target bar detected", "x", m.X, "y", m.Y, "length", m.Length, "thickness", m.Thickness)
		if !a.mutedByUs {
			muted, err := a.audio.Muted(ctx)
			if err != nil {
				return err
			}
			if muted {
				a.log.Debug("audio already muted; ownership not claimed")
				return nil
			}
			if err := a.audio.SetMuted(ctx, true); err != nil {
				return err
			}
			a.mutedByUs = true
			a.mutedAt = time.Now()
			a.log.Info("audio muted", "reason", "target bar detected")
		}
		return nil
	}
	if !a.disappearanceConfirmed() {
		a.log.Debug("target bar temporarily absent", "consecutive_frames", a.missingFrames)
		return nil
	}
	return a.restoreIfAllowed(ctx, true)
}

func (a *App) disappearanceConfirmed() bool {
	if !a.mutedByUs {
		return true
	}
	a.missingFrames++
	return a.missingFrames >= disappearanceConfirmationFrames
}

func (a *App) restoreIfAllowed(ctx context.Context, disappeared bool) error {
	if !a.mutedByUs {
		return nil
	}
	age := time.Since(a.mutedAt)
	if age > a.cfg.RestoreWithin.Duration {
		a.mutedByUs = false
		a.missingFrames = 0
		a.log.Warn("audio left muted", "reason", "restore window expired", "muted_for", age.String())
		return nil
	}
	if err := a.audio.SetMuted(ctx, false); err != nil {
		return fmt.Errorf("restore audio: %w", err)
	}
	a.mutedByUs = false
	a.missingFrames = 0
	a.log.Info("audio restored", "bar_disappeared", disappeared, "muted_for", age.String())
	return nil
}
