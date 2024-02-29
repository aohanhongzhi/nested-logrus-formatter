package formatter

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestRobot(t *testing.T) {
	LogInitRobot(false, true, "test")
	//log.AddHook(NewRobotLogger("app-test"))

	log.Error("测试")

	// Use logrus as normal
	log.WithFields(log.Fields{
		"app": "walrus",
	}).Error("Could not find a bucket")
}
