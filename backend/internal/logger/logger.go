package logger

import (
	"log/slog"
	"os"

	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/config"
	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/logger"

// InitLogger initializes the global structured logger based on the application config.
func InitLogger(cfg config.Config) {
       var logHandler slog.Handler

       level := new(slog.LevelVar)
       switch cfg.LogLevel {
       case "DEBUG":
	       level.Set(slog.LevelDebug)
       case "WARN":
	       level.Set(slog.LevelWarn)
       case "ERROR":
	       level.Set(slog.LevelError)
       default:
	       level.Set(slog.LevelInfo)
       }

       handlerOpts := &slog.HandlerOptions{
	       Level: level,
       }

       var outputWriter = os.Stdout
       // If CloudWatch logging is enabled, use CloudWatchWriter
       if cfg.CloudWatchLogGroup != "" && cfg.CloudWatchLogStream != "" {
	       cw, err := logger.NewCloudWatchWriter(cfg.CloudWatchLogGroup, cfg.CloudWatchLogStream)
	       if err == nil {
		       outputWriter = cw
	       } else {
		       // fallback to stdout, optionally log the error
	       }
       }

       if cfg.LogFormat == "json" {
	       logHandler = slog.NewJSONHandler(outputWriter, handlerOpts)
       } else {
	       logHandler = slog.NewTextHandler(outputWriter, handlerOpts)
       }

       logger := slog.New(logHandler)
       slog.SetDefault(logger)
}
