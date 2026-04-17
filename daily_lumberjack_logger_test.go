package formatter

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDailyLumberjackLoggerRotatesOnDayChange(t *testing.T) {
	t.Cleanup(func() { nowFunc = time.Now })

	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	nowFunc = func() time.Time {
		return time.Date(2024, 1, 1, 10, 0, 0, 0, time.Local)
	}

	logger := newDailyLumberjackLogger(path, 20*1024*1024, 24*time.Hour, 3, false)
	t.Cleanup(func() {
		if err := logger.Close(); err != nil {
			t.Fatalf("close logger failed: %v", err)
		}
	})
	if _, err := logger.Write([]byte("day1\n")); err != nil {
		t.Fatalf("write day1 failed: %v", err)
	}

	nowFunc = func() time.Time {
		return time.Date(2024, 1, 2, 1, 0, 0, 0, time.Local)
	}

	if _, err := logger.Write([]byte("day2\n")); err != nil {
		t.Fatalf("write day2 failed: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read current log failed: %v", err)
	}
	if string(content) != "day2\n" {
		t.Fatalf("current day log mixed with previous day, got: %q", string(content))
	}

	var rotated []string
	for i := 0; i < 20; i++ {
		rotated, _ = filepath.Glob(filepath.Join(dir, "test-*"))
		if len(rotated) > 0 {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	if len(rotated) == 0 {
		t.Fatalf("expected rotated file for previous day, found none")
	}
}

func TestDailyLumberjackLoggerRotatesStaleFileOnStartup(t *testing.T) {
	t.Cleanup(func() { nowFunc = time.Now })

	dir := t.TempDir()
	path := filepath.Join(dir, "stale.log")
	if err := os.WriteFile(path, []byte("yesterday\n"), 0o644); err != nil {
		t.Fatalf("prepare stale log failed: %v", err)
	}
	if err := os.Chtimes(path, time.Date(2024, 2, 1, 23, 59, 0, 0, time.Local), time.Date(2024, 2, 1, 23, 59, 0, 0, time.Local)); err != nil {
		t.Fatalf("set stale log time failed: %v", err)
	}

	nowFunc = func() time.Time {
		return time.Date(2024, 2, 2, 0, 1, 0, 0, time.Local)
	}

	logger := newDailyLumberjackLogger(path, 20*1024*1024, 24*time.Hour, 3, false)
	t.Cleanup(func() {
		if err := logger.Close(); err != nil {
			t.Fatalf("close logger failed: %v", err)
		}
	})
	if _, err := logger.Write([]byte("today\n")); err != nil {
		t.Fatalf("write current day failed: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read current log failed: %v", err)
	}
	if string(content) != "today\n" {
		t.Fatalf("stale content remained after rotation, got: %q", string(content))
	}
}

func TestNewSizeLumberjackLoggerCompressSetting(t *testing.T) {
	compressed := newSizeLumberjackLogger("compressed.log", 20*1024*1024, 24*time.Hour, 3, true)
	if !compressed.Compress {
		t.Fatal("expected compressed logger to enable compression")
	}

	uncompressed := newSizeLumberjackLogger("error.log", 20*1024*1024, 24*time.Hour, 3, false)
	if uncompressed.Compress {
		t.Fatal("expected uncompressed logger to disable compression")
	}
}
