package logging

import (
	"log/slog"
	"os"
)

func Setup() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	slog.SetDefault(logger)
}
