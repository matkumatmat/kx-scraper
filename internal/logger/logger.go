package logger

import (
	"log/slog"
	"os"
)

// Setup menginisialisasi global logger untuk seluruh aplikasi
func Setup(format string) {
	var handler slog.Handler

	// Opsi logger, lu bisa ganti LevelInfo buat production biar log debug mati
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug, 
	}

	// Pilih format output (JSON atau Text)
	if format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	// Bikin instance logger dan set sebagai Global Default
	logger := slog.New(handler)
	slog.SetDefault(logger)
}