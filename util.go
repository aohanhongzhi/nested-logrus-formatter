package formatter

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// 获取当前执行文件的目录，但是IDE下获取当前工程目录。两者都需要兼容。
func GetCurrentPath() string {
	var path string
	ex, err1 := os.Executable()
	if err1 != nil {
		logrus.Error(err1)
	}
	logrus.Debugf("当前程序路径 %v", ex)
	exPath := filepath.Dir(ex) // IDE开发的时候，路径可能不对。生成执行的时候应该是对的。

	logrus.Debugf("当前程序所在目录 %v", exPath)
	pwd, _ := os.Getwd()
	logrus.Debugf("当前程序执行所在目录 %v", pwd)
	if strings.Contains(exPath, "/tmp/fleet") || strings.Contains(exPath, "/tmp/GoLand") || strings.Contains(exPath, "T/GoLand") || strings.Contains(exPath, "\\Temp\\GoLand") || strings.Contains(exPath, "\\tmp\\GoLand") {
		// Goland等IDE调试
		path = pwd
	} else {
		path = exPath
	}
	return path
}

// FIXME: 这里注意日志文件启动路径会不会随着脚本启动的时候执行目录不一样，日志文件存储也不一样。日志不是与可执行文件同一目录，而是与执行启动目录在一起。
func MkLogdir(logPath string) {
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		err1 := os.MkdirAll(logPath, os.ModePerm)
		if err1 != nil {
			logrus.Errorf("%v日志文件夹创建失败%+v", logPath, err1)
			logPath = "." + logPath // 表示建在当前目录下
			if _, err := os.Stat(logPath); os.IsNotExist(err) {
				err1 := os.MkdirAll(logPath, os.ModePerm)
				if err1 != nil {
					logrus.Errorf("当前目录的日志文件夹[%v]创建失败 %+v", logPath, err1)
				} else {
					logrus.Warnf("指定的日志目录%v 无法新建，创建了当前目录下的日志文件夹", logPath)
				}
			}
		}
	}
}
