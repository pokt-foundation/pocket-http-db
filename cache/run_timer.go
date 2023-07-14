package cache

import (
	"time"

	"github.com/sirupsen/logrus"
)

type runTimer struct {
	name      string
	startTime time.Time
	log       *logrus.Logger
}

func NewRunTimer(name string, log *logrus.Logger) *runTimer {

	return &runTimer{
		name:      name,
		startTime: time.Now(),
		log:       log,
	}
}

func (t runTimer) Stop() {
	elapsed := time.Since(t.startTime)
	if elapsed > 0*time.Millisecond {
		t.log.WithFields(logrus.Fields{
			"name":         t.name,
			"milliseconds": elapsed.Milliseconds(),
		}).Info("RunTimer")
	}
}
