package formatter

import (
	"io"
	"time"

	"github.com/sirupsen/logrus"
)

var AppName string
var defaultReserveDuration = time.Duration(72) * time.Hour
var defaultRotationSize int64 = 20 * 1024 * 1024

// 考虑单元测试里面的兼容性，所以新增增加的函数名不一样
func LogInit(noConsole bool) io.Writer {
	return LogrusInit(noConsole, "go-app", ".", logrus.InfoLevel, defaultReserveDuration, defaultRotationSize)
}

// 本配置处理了三个日志输出，1. 控制台（二选一） 2. all.log 所有日志 （二选一） 3. log文件夹下面的分级日志（一定会输出）
// Deprecated
func LogInitRobot(noConsole, robot bool, appName string) io.Writer {
	// 使用 .表示当前路径
	return LogrusInit(noConsole, appName, ".", logrus.InfoLevel, defaultReserveDuration, defaultRotationSize)
}

// 本配置处理了三个日志输出，1. 控制台（二选一） 2. all.log 所有日志 （二选一） 3. log文件夹下面的分级日志（一定会输出）
func LogInitWithName(noConsole bool, appName string) io.Writer {
	// 使用 .表示当前路径
	return LogrusInit(noConsole, appName, ".", logrus.InfoLevel, defaultReserveDuration, defaultRotationSize)
}

func LogInitWithLevel(noConsole bool, appName string, level logrus.Level) io.Writer {
	// 使用 .表示当前路径
	return LogrusInit(noConsole, appName, ".", level, defaultReserveDuration, defaultRotationSize)
}

func LogInitWithParam(noConsole bool, appName string, level logrus.Level, reserveDuration time.Duration, rotationSize int64) io.Writer {
	// 使用 .表示当前路径
	return LogrusInit(noConsole, appName, ".", level, reserveDuration, rotationSize)
}
