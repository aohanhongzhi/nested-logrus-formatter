package formatter

import (
	"io"
	"time"

	"github.com/sirupsen/logrus"
)

var AppName string

// LogBaseDir 保存日志的根目录（由 LogrusInit 赋值）
var LogBaseDir string

// 供扩展能力（如标签日志）复用的全局配置（由 LogrusInit 赋值）
var (
	GlobalReserveDuration       = DefaultReserveDuration
	GlobalRotationSize    int64 = DefaultRotationSize
	GlobalMaxBackups            = 3
	GlobalRotationMode          = DefaultRotationMode
)

const DefaultReserveDuration = time.Duration(72) * time.Hour
const DefaultRotationSize int64 = 20 * 1024 * 1024
const DefaultRotationMode = RotationModeTime

type RotationMode string

const (
	RotationModeTime RotationMode = "time"
	RotationModeSize RotationMode = "size"
)

// 考虑单元测试里面的兼容性，所以新增增加的函数名不一样
func LogInit(noConsole bool) io.Writer {
	return LogrusInit(noConsole, "go-app", ".", logrus.InfoLevel, DefaultRotationMode, DefaultReserveDuration, DefaultRotationSize, 3)
}

// 本配置处理了三个日志输出，1. 控制台（二选一） 2. all.log 所有日志 （二选一） 3. log文件夹下面的分级日志（一定会输出）
// Deprecated
func LogInitRobot(noConsole, robot bool, appName string) io.Writer {
	// 使用 .表示当前路径
	return LogrusInit(noConsole, appName, ".", logrus.InfoLevel, DefaultRotationMode, DefaultReserveDuration, DefaultRotationSize, 3)
}

// 本配置处理了三个日志输出，1. 控制台（二选一） 2. all.log 所有日志 （二选一） 3. log文件夹下面的分级日志（一定会输出）
func LogInitWithName(noConsole bool, appName string) io.Writer {
	// 使用 .表示当前路径
	return LogrusInit(noConsole, appName, ".", logrus.InfoLevel, DefaultRotationMode, DefaultReserveDuration, DefaultRotationSize, 3)
}

func LogInitWithLevel(noConsole bool, appName string, level logrus.Level) io.Writer {
	// 使用 .表示当前路径
	return LogrusInit(noConsole, appName, ".", level, DefaultRotationMode, DefaultReserveDuration, DefaultRotationSize, 3)
}

func LogInitWithParam(noConsole bool, appName string, level logrus.Level, reserveDuration time.Duration, rotationSize int64) io.Writer {
	// 使用 .表示当前路径
	return LogrusInit(noConsole, appName, ".", level, DefaultRotationMode, reserveDuration, rotationSize, 3)
}

func LogInitWithMaxBackup(noConsole bool, appName string, level logrus.Level, reserveDuration time.Duration, rotationSize int64, maxBackups int) io.Writer {
	// 使用 .表示当前路径
	return LogrusInit(noConsole, appName, ".", level, DefaultRotationMode, reserveDuration, rotationSize, maxBackups)
}

func LogInitWithDir(noConsole bool, appName string, logDir string, level logrus.Level, rotationMode RotationMode, reserveDuration time.Duration, rotationSize int64, maxBackups int) io.Writer {
	if logDir == "" {
		logDir = GetCurrentPath()
	}
	return LogrusInit(noConsole, appName, logDir, level, rotationMode, reserveDuration, rotationSize, maxBackups)
}
