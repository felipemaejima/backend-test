package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Level  string
	Format string
	File   string
}

func New(cfg Config) (*slog.Logger, func() error, error) {
	output := io.Writer(os.Stdout)
	closeFn := func() error { return nil }

	if cfg.File != "" {
		file, err := openFile(cfg.File)
		if err != nil {
			return nil, nil, err
		}
		output = io.MultiWriter(os.Stdout, file)
		closeFn = file.Close
	}

	options := &slog.HandlerOptions{Level: parseLevel(cfg.Level)}

	var handler slog.Handler
	if strings.EqualFold(cfg.Format, "text") {
		handler = slog.NewTextHandler(output, options)
	} else {
		handler = slog.NewJSONHandler(output, options)
	}

	return slog.New(handler), closeFn, nil
}

func openFile(path string) (*os.File, error) {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("criando diretório de log %q: %w", dir, err)
		}
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("abrindo arquivo de log %q: %w", path, err)
	}
	return file, nil
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
