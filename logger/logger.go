package logger

import (
	"fmt"
	"log"

	"go.uber.org/zap"
)

// NewLogger creates the logger.
func NewLogger(level string) (logger *zap.Logger, err error) {
	l := zap.InfoLevel
	switch level {
	case "debug":
		l = zap.DebugLevel
	case "info":
		l = zap.InfoLevel
	case "":
		l = zap.InfoLevel
	case "error":
		l = zap.ErrorLevel
	case "warn":
		l = zap.WarnLevel
	case "dpanic":
		l = zap.DPanicLevel
	case "panic":
		l = zap.PanicLevel
	case "fatal":
		l = zap.FatalLevel
	default:
		log.Fatal(fmt.Sprintf("Failed to parse log level %s", level))
	}
	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(l),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
	logger, err = cfg.Build()
	if err != nil {
		return
	}
	defer logger.Sync()
	return
}
