package formatter_test

import (
	"testing"

	formatter "github.com/aohanhongzhi/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
)

func LogInitRobot(t *testing.T) {
	formatter.LogInitRobot(true, true, "test")
	log.Info("Testing")
	log.Error("Testing")
}

func TestLog(t *testing.T) {
	formatter.LogInit(true)
	log.Info("Testing")
	log.Error("Testing")
}
