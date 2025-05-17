package formatter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/natefinch/lumberjack"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

// 支持日志存放位置
func LogrusInit(noConsole bool, appName, dir string, level logrus.Level, reserveDuration time.Duration, rotationSize int64) io.Writer {
	// 设置时区为东八区
	os.Setenv("TZ", "Asia/Shanghai")
	AppName = appName
	// 参考文章 https://juejin.cn/post/7026912807333888014
	logPath := filepath.Join(dir, "/log")
	debugLogPath := filepath.Join(dir, "/log/debug/")
	infoLogPath := filepath.Join(dir, "/log/info/")
	warnLogPath := filepath.Join(dir, "/log/warn/")
	errorLogPath := filepath.Join(dir, "/log/error/")
	panicLogPath := filepath.Join(dir, "/log/panic/")

	// TODO 等待有对应日志的时候再生成对应文件夹比较好
	MkLogdir(logPath)
	MkLogdir(debugLogPath)
	MkLogdir(infoLogPath)
	MkLogdir(warnLogPath)
	MkLogdir(errorLogPath)
	MkLogdir(panicLogPath)

	logFileName := filepath.Join(logPath, "all")
	debugLogFileName := filepath.Join(debugLogPath, "debug")
	infoLogFileName := filepath.Join(infoLogPath, "info")
	warnlogFileName := filepath.Join(warnLogPath, "warn")
	errorlogFileName := filepath.Join(errorLogPath, "error")
	paniclogFileName := filepath.Join(panicLogPath, "panic")

	// 设置项目默认日志级别
	logrus.SetLevel(level)

	logrus.SetReportCaller(true)

	fileFormatter := &Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		NoColors:        false, // 服务器查看文件有颜色
		HideKeys:        true,
		NoFieldsSpace:   false,
		FieldsOrder:     []string{"component", "category", "req"},
		CustomCallerFormatter: func(f *runtime.Frame) string {
			file, line := f.File, f.Line
			if strings.HasPrefix(f.Function, "github.com/aohanhongzhi/gormv2-logrus") {
				// gorm框架日志特殊处理
				_, file1, line1, ok := runtime.Caller(14)
				if !ok {
					logrus.Errorf("获取行号失败 %v,%v", file1, line1)
				}
				//sprintf := fmt.Sprintf(" fileFormatter (%s:%d) => (%s:%d)", file1, line1, file, line)
				//println(sprintf)
				return fmt.Sprintf(" (%s:%d)  => (%s:%d) ", file1, line1, file, line)
			}

			return fmt.Sprintf(" (%s:%d)", file, line)
		},
	}

	stdoutFormatter := &Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		HideKeys:        true,
		NoFieldsSpace:   false,
		FieldsOrder:     []string{"component", "category", "req"},
		CustomCallerFormatter: func(f *runtime.Frame) string {
			file, line := f.File, f.Line
			if strings.HasPrefix(f.Function, "github.com/aohanhongzhi/gormv2-logrus") {
				_, file1, line1, ok := runtime.Caller(11)
				if !ok {
					logrus.Errorf("获取行号失败 %v,%v", file1, line1)
				}
				//sprintf := fmt.Sprintf(" stdoutFormatter (%s:%d) => (%s:%d)", file1, line1, file, line)
				//println(sprintf)
				return fmt.Sprintf(" %s:%d  %s:%d  ", file1, line1, file, line)
			}
			return fmt.Sprintf(" %s:%d", f.File, f.Line)
		},
	}

	// 下面配置日志大小达到10M就会生成一个新文件，保留最近 3 天的日志文件，多余的自动清理掉。
	// 参考文章 https://blog.csdn.net/qq_42119514/article/details/121372416
	writer, _ := rotatelogs.New(
		logFileName+"-%Y%m%d%H%M.log",
		//rotatelogs.WithLinkName(logFilePath),
		rotatelogs.WithMaxAge(reserveDuration), //保留最近 3 天的日志文件，多余的自动清理掉
		//rotatelogs.WithRotationTime(time.Duration(6)*time.Hour), // 每隔 6小时轮转一个新文件
		rotatelogs.WithRotationSize(rotationSize), //设置10MB大小,当大于这个容量时，创建新的日志文件
	)

	debugWriter, _ := rotatelogs.New(
		debugLogFileName+"-%Y%m%d%H%M.log",
		//rotatelogs.WithLinkName(logFilePath),
		rotatelogs.WithMaxAge(time.Duration(72)*time.Hour), //保留最近 3 天的日志文件，多余的自动清理掉
		//rotatelogs.WithRotationTime(time.Duration(6)*time.Hour), // 每隔 6小时轮转一个新文件
		rotatelogs.WithRotationSize(rotationSize), //设置10MB大小,当大于这个容量时，创建新的日志文件
	)

	infoWriter, _ := rotatelogs.New(
		infoLogFileName+"-%Y%m%d%H%M.log",
		//rotatelogs.WithLinkName(logFilePath),
		rotatelogs.WithMaxAge(time.Duration(72)*time.Hour), //保留最近 3 天的日志文件，多余的自动清理掉
		//rotatelogs.WithRotationTime(time.Duration(6)*time.Hour), // 每隔 6小时轮转一个新文件
		rotatelogs.WithRotationSize(rotationSize), //设置10MB大小,当大于这个容量时，创建新的日志文件
	)

	warnWriter, _ := rotatelogs.New(
		warnlogFileName+"-%Y%m%d%H%M.log",
		//rotatelogs.WithLinkName(logFilePath),
		rotatelogs.WithMaxAge(time.Duration(72)*time.Hour), //保留最近 3 天的日志文件，多余的自动清理掉
		//rotatelogs.WithRotationTime(time.Duration(6)*time.Hour), // 每隔 6小时轮转一个新文件
		rotatelogs.WithRotationSize(rotationSize), //设置10MB大小,当大于这个容量时，创建新的日志文件
	)

	errorWriter, _ := rotatelogs.New(
		errorlogFileName+"-%Y%m%d%H%M.log",
		//rotatelogs.WithLinkName(logFilePath),
		rotatelogs.WithMaxAge(time.Duration(72)*time.Hour), //保留最近 3 天的日志文件，多余的自动清理掉
		//rotatelogs.WithRotationTime(time.Duration(6)*time.Hour), // 每隔 6小时轮转一个新文件
		rotatelogs.WithRotationSize(rotationSize), //设置10MB大小,当大于这个容量时，创建新的日志文件
	)

	panicWriter, _ := rotatelogs.New(
		paniclogFileName+"-%Y%m%d%H%M.log",
		//rotatelogs.WithLinkName(logFilePath),
		rotatelogs.WithMaxAge(time.Duration(72)*time.Hour), //保留最近 3 天的日志文件，多余的自动清理掉
		//rotatelogs.WithRotationTime(time.Duration(6)*time.Hour), // 每隔 6小时轮转一个新文件
		rotatelogs.WithRotationSize(rotationSize), //设置10MB大小,当大于这个容量时，创建新的日志文件
	)

	writers := []io.Writer{writer, errorWriter}

	//同时写到两个文件里
	allLevelWriter := io.MultiWriter(writers...)

	lfHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: writer, // 为不同级别设置不同的输出目的
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: allLevelWriter,
		logrus.PanicLevel: allLevelWriter,
		logrus.FatalLevel: allLevelWriter,
	}, fileFormatter)
	logrus.AddHook(lfHook) // 输出到log文件夹（一定会输出）

	debuglfHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: debugWriter,
	}, fileFormatter)
	logrus.AddHook(debuglfHook) // 输出到log文件夹（一定会输出）

	infolfHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.InfoLevel: infoWriter,
	}, fileFormatter)
	logrus.AddHook(infolfHook) // 输出到log文件夹（一定会输出）

	warnlfHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.WarnLevel: warnWriter,
	}, fileFormatter)
	logrus.AddHook(warnlfHook) // 输出到log文件夹（一定会输出）

	paniclfHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.PanicLevel: panicWriter,
	}, fileFormatter)
	logrus.AddHook(paniclfHook) // 输出到log文件夹（一定会输出）

	// 下面是另一个日志文件处理方式

	fileWriter := &lumberjack.Logger{
		Filename:   "all.log",
		MaxSize:    int(rotationSize) / (1024 * 1024), // megabytes
		MaxBackups: 2,
		MaxAge:     int(reserveDuration.Hours() / 24), //days
		Compress:   true,                              // disabled by default
	}

	var multiWriter io.Writer
	if noConsole {
		multiWriter = io.MultiWriter(fileWriter) // 覆盖上面的控制台输出
		logrus.SetFormatter(fileFormatter)
	} else {
		// 控制台和文件都有，因为有时候控制台看起来麻烦，一旦重启就没了，所以还是需要持久化存储
		multiWriter = io.MultiWriter(os.Stdout, fileWriter) // 控制台+文件持久化
		logrus.SetFormatter(stdoutFormatter)
	}
	logrus.SetOutput(multiWriter)
	return multiWriter
}
