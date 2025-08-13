package formatter_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	formatter "github.com/aohanhongzhi/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
)

func TestTagRouterCreatesTagFile(t *testing.T) {
	// 初始化日志（控制台+文件），并启用 Tag 路由
	formatter.LogInitWithLevel(false, "", log.DebugLevel)
	formatter.EnableTagFileRouter()

	// 写入带有 tag 的日志
	log.WithField("tag", "orders").Info("created order #1")
	log.WithField("tag", "orders").Error("order error")

	log.WithField("tag", "book").Error("book error")

	// 等待异步 hook 处理完
	formatter.FlushAsyncHooks()
	// 保险起见，稍作等待
	time.Sleep(50 * time.Millisecond)

	// 断言文件存在
	path := filepath.Join(".", "log", "tag", "orders.log")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expect tag log file exists, got error: %v", err)
	}
}
