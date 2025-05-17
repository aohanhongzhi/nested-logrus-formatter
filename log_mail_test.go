package formatter

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestSendMail(t *testing.T) {
	hook1, err := NewMailAuthHook("core", "smtp.qq.com", 25, "aohanhongzhi@qq.com", "3227556776@qq.com", "aohanhongzhi@qq.com", "password")
	if err == nil {
		logrus.AddHook(hook1)
	}
	logrus.Errorf("错误日志发送邮件")
}
