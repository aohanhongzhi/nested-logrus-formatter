package formatter

import (
	"testing"
	"time"
)

func TestFeishuRobot(t *testing.T) {
	FeishuRobotDetail("第几行", "飞书测试")
	time.Sleep(5 * time.Second)
}
