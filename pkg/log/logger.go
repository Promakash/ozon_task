package log

import (
	"io"
	"log"
	"log/slog"
	"os"
	"time"
)

type Config struct {
	Level     string `yaml:"level" env-default:"info"`
	Format    string `yaml:"format" env-default:"text"`
	Directory string `yaml:"directory" env-default:""`
}

func NewLogger(cfg Config) (*slog.Logger, *os.File) {
	var handler slog.Handler

	var logLevel slog.Level
	switch cfg.Level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	needFile := len(cfg.Directory) != 0

	file, err := createLogFile(cfg.Directory)
	if err != nil && needFile {
		log.Fatal("error while creating log file")
	}

	var writer io.Writer
	if needFile {
		writer = io.MultiWriter(os.Stdout, file)
	} else {
		writer = os.Stdout
	}

	switch cfg.Format {
	case "text":
		handler = slog.NewTextHandler(writer, &slog.HandlerOptions{Level: logLevel})
	default:
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: logLevel})
	}

	return slog.New(handler), file
}

func createLogFile(directory string) (*os.File, error) {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		if err = os.MkdirAll(directory, 0755); err != nil {
			return nil, err
		}
	}

	filePath := directory + "/" + time.Now().String() + ".log"

	logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return logFile, nil
}

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func Fatal(l *slog.Logger, msg string, err error) {
	l.Error(msg, Err(err))
	os.Exit(1)
}
