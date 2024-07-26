package formatter

import (
    log "github.com/sirupsen/logrus"
    "os"
    "path/filepath"
    "strings"
)

func GetCurrentPath() string {
	var path string
	ex, err1 := os.Executable()
	if err1 != nil {
		log.Error(err1)
	}
	log.Debugf("当前程序路径 %v", ex)
	exPath := filepath.Dir(ex) // IDE开发的时候，路径可能不对。生成执行的时候应该是对的。
	
	log.Debugf("当前程序所在目录 %v", exPath)
	pwd, _ := os.Getwd()
	log.Debugf("当前程序执行所在目录 %v", pwd)
	if strings.Contains(exPath, "/tmp/fleet") || strings.Contains(exPath, "/tmp/GoLand") || strings.Contains(exPath, "T/GoLand") || strings.Contains(exPath, "\\Temp\\GoLand") || strings.Contains(exPath, "\\tmp\\GoLand") {
		// Goland等IDE调试
		path = pwd
	} else {
		path = exPath
	}
	return path
}
