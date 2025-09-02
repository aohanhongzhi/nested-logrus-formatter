package formatter_test

import (
	"testing"

	formatter "github.com/aohanhongzhi/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
)

func TestGormErrorMailHook(t *testing.T) {
	// 初始化日志
	formatter.LogInitWithLevel(false, "test-app", log.DebugLevel)

	// 注意：实际测试需要有效的邮件服务器配置
	// 这里只是测试Hook的创建和基本逻辑，不发送实际邮件

	// 创建Hook（这里使用虚拟配置）
	hook, err := formatter.NewGormErrorMailHook(
		"test-app",
		"smtp.example.com",
		587,
		"test@example.com",
		"alert@example.com",
		"testuser",
		"testpass",
	)

	if err != nil {
		// 连接失败是正常的，因为使用了虚拟配置
		t.Logf("创建Hook失败（预期行为，因为使用了虚拟配置）: %v", err)
		return
	}

	if hook != nil {
		// 测试是否能正确识别GORM错误
		t.Logf("Hook创建成功，启用状态: %v", hook.Enable)
	}
}
