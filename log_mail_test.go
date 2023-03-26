package formatter

import (
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestSendMail(t *testing.T) {
	hook1, err := NewMailAuthHook("core", "smtp.qq.com", 25, "aohanhongzhi@qq.com", "3227556776@qq.com", "aohanhongzhi@qq.com", "password")
	if err == nil {
		log.AddHook(hook1)
	}
	log.Errorf("错误日志发送邮件")
}
