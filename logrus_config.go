package formatter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/natefinch/lumberjack"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
)

var AppName string

// 考虑单元测试里面的兼容性，所以新增增加的函数名不一样
func LogInit(noConsole bool) io.Writer {
	return LogInitRobot(noConsole, false, "go-app")
}

// 本配置处理了三个日志输出，1. 控制台（二选一） 2. all.log 所有日志 （二选一） 3. log文件夹下面的分级日志（一定会输出）
func LogInitRobot(noConsole, robot bool, appName string) io.Writer {
	AppName = appName
	// 参考文章 https://juejin.cn/post/7026912807333888014
	logPath := "./log"
	errorLogPath := "./log/error/"
	// FIXME: 这里注意日志文件启动路径会不会随着脚本启动的时候执行目录不一样，日志文件存储也不一样。日志不是与可执行文件同一目录，而是与执行启动目录在一起。
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		err1 := os.Mkdir(logPath, os.ModePerm)
		if err1 != nil {
			log.Errorf("日志文件夹创建失败 %+v", err1)
		}
	}
	if _, err := os.Stat(errorLogPath); os.IsNotExist(err) {
		err1 := os.Mkdir(errorLogPath, os.ModePerm)
		if err1 != nil {
			log.Errorf("Error日志文件夹创建失败%+v", err1)
		}
	}
	logFilePath := filepath.Join(logPath, "go")
	errorlogFilePath := filepath.Join(errorLogPath, "error")

	// 设置项目默认日志级别
	log.SetLevel(log.InfoLevel)

	log.SetReportCaller(true)

	fileFormatter := &Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		NoColors:        false, // 服务器查看文件有颜色
		HideKeys:        true,
		NoFieldsSpace:   false,
		FieldsOrder:     []string{"component", "category", "req"},
		CustomCallerFormatter: func(f *runtime.Frame) string {
			return fmt.Sprintf(" (%s:%d)", f.File, f.Line)
		},
	}
	stdoutFormatter := &Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		HideKeys:        true,
		NoFieldsSpace:   false,
		FieldsOrder:     []string{"component", "category", "req"},
		CustomCallerFormatter: func(f *runtime.Frame) string {
			return fmt.Sprintf(" %s:%d", f.File, f.Line)
		},
	}
	// 下面配置日志大小达到10M就会生成一个新文件，保留最近 3 天的日志文件，多余的自动清理掉。
	// 参考文章 https://blog.csdn.net/qq_42119514/article/details/121372416
	writer, _ := rotatelogs.New(
		logFilePath+"-%Y%m%d%H%M.log",
		//rotatelogs.WithLinkName(logFilePath),
		rotatelogs.WithMaxAge(time.Duration(72)*time.Hour), //保留最近 3 天的日志文件，多余的自动清理掉
		//rotatelogs.WithRotationTime(time.Duration(6)*time.Hour), // 每隔 6小时轮转一个新文件
		rotatelogs.WithRotationSize(10*1024*1024), //设置10MB大小,当大于这个容量时，创建新的日志文件
	)

	errorWriter, _ := rotatelogs.New(
		errorlogFilePath+"-%Y%m%d%H%M.log",
		//rotatelogs.WithLinkName(logFilePath),
		rotatelogs.WithMaxAge(time.Duration(72)*time.Hour), //保留最近 3 天的日志文件，多余的自动清理掉
		//rotatelogs.WithRotationTime(time.Duration(6)*time.Hour), // 每隔 6小时轮转一个新文件
		rotatelogs.WithRotationSize(10*1024*1024), //设置10MB大小,当大于这个容量时，创建新的日志文件
	)
	writers := []io.Writer{
		writer,
		errorWriter}
	//同时写到两个文件里
	allLevelWriter := io.MultiWriter(writers...)

	lfHook := lfshook.NewHook(lfshook.WriterMap{
		log.DebugLevel: writer, // 为不同级别设置不同的输出目的
		log.InfoLevel:  writer,
		log.WarnLevel:  writer,
		log.ErrorLevel: allLevelWriter,
		log.FatalLevel: allLevelWriter,
		log.PanicLevel: allLevelWriter,
	}, fileFormatter)
	log.AddHook(lfHook) // 输出到log文件夹（一定会输出）

	fileWriter := &lumberjack.Logger{
		Filename:   "all.log",
		MaxSize:    50, // megabytes
		MaxBackups: 2,
		MaxAge:     2,    //days
		Compress:   true, // disabled by default
	}
	multiWriter := io.MultiWriter(os.Stdout) // 输出到控制台
	log.SetFormatter(stdoutFormatter)
	if noConsole {
		multiWriter = io.MultiWriter(fileWriter) // 覆盖上面的控制台输出
		log.SetFormatter(fileFormatter)
	}
	log.SetOutput(multiWriter)
	// log.SetOutput(os.Stdout) // 直接输出控制台

	// gin的日志接管
	// gin.DefaultWriter = multiWriter

	if robot {
		log.AddHook(NewRobotLogger(appName))
	}

	return multiWriter
}
