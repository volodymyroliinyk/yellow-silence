package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/volodymyr/yellow-silence/internal/app"
	"github.com/volodymyr/yellow-silence/internal/config"
)

const version = "0.1.0"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "yellow-silence:", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		usage()
		return fmt.Errorf("command is required")
	}
	switch os.Args[1] {
	case "run", "check":
		path, err := configPath(os.Args[2:])
		if err != nil {
			return err
		}
		cfg, err := config.Load(path)
		if err != nil {
			return err
		}
		if os.Args[1] == "check" {
			fmt.Printf("configuration is valid: %s\n", path)
			return nil
		}
		logger := newLogger(cfg.LogLevel)
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		return app.New(cfg, logger).Run(ctx)
	case "init-config":
		path, err := configPath(os.Args[2:])
		if err != nil {
			return err
		}
		return config.WriteExample(path)
	case "version", "--version", "-v":
		fmt.Println(version)
		return nil
	case "help", "--help", "-h":
		usage()
		return nil
	default:
		usage()
		return fmt.Errorf("unknown command %q", os.Args[1])
	}
}

func configPath(args []string) (string, error) {
	path := config.DefaultPath()
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--config", "-c":
			if i+1 >= len(args) {
				return "", fmt.Errorf("%s requires a path", args[i])
			}
			i++
			path = args[i]
		default:
			return "", fmt.Errorf("unknown option %q", args[i])
		}
	}
	return path, nil
}

func newLogger(level string) *slog.Logger {
	var l slog.Level
	_ = l.UnmarshalText([]byte(level))
	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: l}))
}

func usage() {
	fmt.Fprintln(os.Stderr, `Usage: yellow-silence <command> [options]

Commands:
  run          monitor the screen and control audio
  check        validate configuration
  init-config  write a secure example configuration
  version      print version

Options:
  -c, --config PATH  configuration file path`)
}
