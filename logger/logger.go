// Package logger is a shared logger for goemu components
package logger

import (
	"log/slog"
	"os"
	"sync"
)

var (
	log  *slog.Logger = nil
	once sync.Once
)

func New() *slog.Logger {
	once.Do(
		func() {
			log = slog.New(
				slog.NewTextHandler(
					os.Stderr, &slog.HandlerOptions{
						Level: slog.LevelDebug,
					},
				))
		},
	)
	return log
}
