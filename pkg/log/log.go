package log

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"

	"backend_go/pkg/format_errors"
)

type Log struct {
	Level        string `mapstructure:"level"`
	IsLogInFiles bool   `mapstructure:"is_log_in_files"`
	Dir          string `mapstructure:"dir"`
	MaxSize      int    `mapstructure:"max_size"`
	MaxBackups   int    `mapstructure:"max_backups"`
	MaxAge       int    `mapstructure:"max_age"`
	Compress     bool   `mapstructure:"compress"`
}

var Defaults = map[string]any{
	"log.level":           "trace",
	"log.is_log_in_files": true,
	"log.dir":             ".logs",
	"log.max_size":        5,
	"log.max_backups":     4,
	"log.max_age":         360,
	"log.compress":        false,
}

func GetLogger(cfg *Log) *zerolog.Logger {
	var (
		logLevel zerolog.Level
		writer   io.Writer
	)
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "2006-01-02 15:04:05",
	}
	if cfg.IsLogInFiles {
		file := filepath.Join(filepath.Clean(cfg.Dir), "app.log")
		logDir := filepath.Dir(file)
		if logDir != "." && logDir != "" {
			if err := os.MkdirAll(logDir, 0o755); err != nil {
				fmt.Fprintf(os.Stderr, "failed to create log directory %s: %v\n", logDir, err)
			}
		}
		fileLogger := &lumberjack.Logger{
			Filename:   file,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
		fileConsoleWriter := zerolog.ConsoleWriter{
			Out:        fileLogger,
			NoColor:    true,
			TimeFormat: "2006-01-02 15:04:05",
		}
		writer = io.MultiWriter(consoleWriter, fileConsoleWriter)
	} else {
		writer = consoleWriter
	}
	if lvl, err := zerolog.ParseLevel(cfg.Level); err == nil {
		logLevel = lvl
	} else {
		logLevel = zerolog.InfoLevel
	}
	logger := zerolog.New(writer).
		Level(logLevel).
		With().
		Timestamp().
		Caller().
		Logger()
	return &logger
}

func StreamLogs(w io.Writer, dir string) error {
	bw := bufio.NewWriter(w)
	defer bw.Flush()
	return filepath.Walk(filepath.Clean(dir), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return format_errors.Wrap(err, "ошибка логирования")
		}
		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			rel = info.Name()
		}

		header := fmt.Sprintf("\n--------------------------------------------------\nFile: %s\n--------------------------------------------------\n", rel)
		if _, err := bw.WriteString(header); err != nil {
			return format_errors.Wrap(err, "ошибка логирования")
		}

		f, err := os.Open(path)
		if err != nil {
			return format_errors.Wrap(err, "ошибка логирования")
		}
		defer f.Close()

		if _, err := io.Copy(bw, f); err != nil {
			return format_errors.Wrap(err, "ошибка логирования")
		}
		return nil
	})
}
