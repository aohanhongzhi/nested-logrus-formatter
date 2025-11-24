package formatter

import (
	"fmt"
	"os"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// nowFunc is used for testability; in production it defaults to time.Now.
var nowFunc = time.Now

type dailyLumberjackLogger struct {
	logger     *lumberjack.Logger
	mu         sync.Mutex
	currentDay string
}

func newDailyLumberjackLogger(filename string, rotationSize int64, reserveDuration time.Duration, maxBackups int) *dailyLumberjackLogger {
	return &dailyLumberjackLogger{
		logger: &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    int(rotationSize) / (1024 * 1024),
			MaxBackups: maxBackups,
			MaxAge:     int(reserveDuration.Hours() / 24),
			Compress:   true,
			LocalTime:  true,
		},
		currentDay: detectFileDay(filename),
	}
}

func detectFileDay(filename string) string {
	info, err := os.Stat(filename)
	if err != nil {
		return dayString(nowFunc())
	}
	return dayString(info.ModTime())
}

func dayString(t time.Time) string {
	return t.In(time.Local).Format("2006-01-02")
}

func (l *dailyLumberjackLogger) Write(p []byte) (int, error) {
	now := nowFunc()
	l.rotateIfNeeded(now)
	return l.logger.Write(p)
}

func (l *dailyLumberjackLogger) rotateIfNeeded(now time.Time) {
	day := dayString(now)
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.currentDay == day {
		return
	}
	if err := l.logger.Rotate(); err != nil {
		fmt.Fprintf(os.Stderr, "lumberjack daily rotate failed: %v\n", err)
	}
	l.currentDay = day
}

func (l *dailyLumberjackLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.logger.Close()
}
