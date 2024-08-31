package setup

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

func SetupLog() *slog.Logger {

	Logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))

	return Logger
}
