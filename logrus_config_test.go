package formatter

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	m.Run()
	LogInitWithLevel(false, "", log.TraceLevel)
	log.Debug("Testing")
	log.Info("Testing")
	log.Warn("Testing")
	log.Error("Testing")
	// log.Panic("Testing")
	// log.Fatal("Testing")

}
