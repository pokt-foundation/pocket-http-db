package cache

import (
	"time"

	"log/slog"

	"github.com/pokt-foundation/utils-go/logger"
)

type runTimer struct {
	name      string
	startTime time.Time
	logger    *logger.Logger
}

func NewRunTimer(name string, logger *logger.Logger) *runTimer {
	return &runTimer{
		name:      name,
		startTime: time.Now(),
		logger:    logger,
	}
}

func (t runTimer) Stop() {
	elapsed := time.Since(t.startTime)
	if elapsed > 0*time.Millisecond {
		t.logger.Info("RunTimer",
			slog.String("name", t.name),
			slog.Int64("milliseconds", elapsed.Milliseconds()),
		)
	}
}
