package formatter_test

import (
	"testing"

	formatter "github.com/aohanhongzhi/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
)

func TestLog(t *testing.T) {
	formatter.LogInit(true, true)
	log.Info("Testing")
	log.Error("Testing")
}
