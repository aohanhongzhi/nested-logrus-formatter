package formatter_test

import (
	"testing"

	formatter "github.com/aohanhongzhi/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
)

func TestLogInitRobot(t *testing.T) {
	formatter.LogInitRobot(true, true, "test")
	log.SetLevel(log.DebugLevel)
	log.Info("Testing")
	log.Error("Testing")
	log.Warn("Testing")
	log.Debug("Testing")
}

func TestLog(t *testing.T) {
	formatter.LogInit(false)
	log.Info("Testing")
	log.Error("Testing")
}
