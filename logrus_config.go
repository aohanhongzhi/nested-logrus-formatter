package formatter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

// AsyncHook 是一个异步的logrus hook，用于避免日志写入阻塞主程序
type AsyncHook struct {
	hook     logrus.Hook
	levels   []logrus.Level
	ch       chan *logrus.Entry
	wg       sync.WaitGroup
	mu       sync.Mutex
	bufSize  int
	shutdown bool
}

// NewAsyncHook 创建一个新的AsyncHook
func NewAsyncHook(hook logrus.Hook, bufSize int) *AsyncHook {
	levels := hook.Levels()
	h := &AsyncHook{
		hook:    hook,
		levels:  levels,
		ch:      make(chan *logrus.Entry, bufSize),
		bufSize: bufSize,
	}

	h.wg.Add(1)
	go h.processLogs()
	return h
}

// Levels 返回此Hook处理的日志级别
func (h *AsyncHook) Levels() []logrus.Level {
	return h.levels
}

// Fire 实现Hook接口，将日志条目发送到异步处理通道
func (h *AsyncHook) Fire(entry *logrus.Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.shutdown {
		return nil
	}

	// 制作日志条目的副本，因为entry在Fire返回后可能会被复用
	copiedEntry := &logrus.Entry{
		Logger:  entry.Logger,
		Data:    make(logrus.Fields, len(entry.Data)),
		Time:    entry.Time,
		Level:   entry.Level,
		Message: entry.Message,
	}

	for k, v := range entry.Data {
		copiedEntry.Data[k] = v
	}

	if entry.Caller != nil {
		caller := *(entry.Caller)
		copiedEntry.Caller = &caller
	}

	// 非阻塞发送，如果通道已满则丢弃
	select {
	case h.ch <- copiedEntry:
		// 成功发送
	default:
		// 通道已满，丢弃此日志（可选择记录警告消息）
		// fmt.Fprintf(os.Stderr, "AsyncHook channel is full, dropping log entry\n")
	}

	return nil
}

// processLogs 在后台goroutine中处理日志条目
func (h *AsyncHook) processLogs() {
	defer h.wg.Done()

	for entry := range h.ch {
		if err := h.hook.Fire(entry); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing log in AsyncHook: %v\n", err)
		}
	}
}

// Flush 等待所有日志条目被处理
func (h *AsyncHook) Flush() {
	h.mu.Lock()
	if h.shutdown {
		h.mu.Unlock()
		return
	}
	h.shutdown = true
	close(h.ch)
	h.mu.Unlock()
	h.wg.Wait()
}

// 支持日志存放位置
func LogrusInit(noConsole bool, appName, dir string, level logrus.Level, reserveDuration time.Duration, rotationSize int64, maxBackups int) io.Writer {
	// 设置时区为东八区
	os.Setenv("TZ", "Asia/Shanghai")
	AppName = appName
	// 自动解析目录：生产环境取可执行文件所在目录；IDE 下取工程工作目录
	if dir == "" {
		dir = GetCurrentPath()
	}
	LogBaseDir = dir
	GlobalReserveDuration = reserveDuration
	GlobalRotationSize = rotationSize
	GlobalMaxBackups = maxBackups
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
					return fmt.Sprintf(" (%s:%d) ", file, line)
				} else {
					return fmt.Sprintf(" (%s:%d)  => (%s:%d) ", file1, line1, file, line)
				}
				//sprintf := fmt.Sprintf(" fileFormatter (%s:%d) => (%s:%d)", file1, line1, file, line)
				//println(sprintf)
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

	writer := newDailyLumberjackLogger(logFileName+".log", rotationSize, reserveDuration, maxBackups)

	// 下面配置日志大小达到10M就会生成一个新文件，保留最近 3 天的日志文件，多余的自动清理掉。 实际上没有清理
	// 参考文章 https://blog.csdn.net/qq_42119514/article/details/121372416
	debugWriter := newDailyLumberjackLogger(debugLogFileName+".log", rotationSize, reserveDuration, maxBackups)

	infoWriter := newDailyLumberjackLogger(infoLogFileName+".log", rotationSize, reserveDuration, maxBackups)

	warnWriter := newDailyLumberjackLogger(warnlogFileName+".log", rotationSize, reserveDuration, maxBackups)

	errorWriter := newDailyLumberjackLogger(errorlogFileName+".log", rotationSize, reserveDuration, maxBackups)

	panicWriter := newDailyLumberjackLogger(paniclogFileName+".log", rotationSize, reserveDuration, maxBackups)

	writers := []io.Writer{writer, errorWriter}

	//同时写到两个文件里
	allLevelWriter := io.MultiWriter(writers...)

	lfHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.TraceLevel: writer, // 为不同级别设置不同的输出目的
		logrus.DebugLevel: writer, // 为不同级别设置不同的输出目的
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: allLevelWriter,
		logrus.PanicLevel: allLevelWriter,
		logrus.FatalLevel: allLevelWriter,
	}, fileFormatter)
	logrus.AddHook(lfHook) // 输出到log文件夹（一定会输出）

	// 为Debug级别使用异步Hook，避免阻塞主程序
	debuglfHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: debugWriter,
	}, fileFormatter)
	// 创建异步Hook，缓冲区大小为1000条日志
	asyncDebugHook := NewAsyncHook(debuglfHook, 1000)
	logrus.AddHook(asyncDebugHook)    // 输出到log文件夹（以异步方式）
	RegisterAsyncHook(asyncDebugHook) // 注册到全局异步Hook列表中，确保程序退出时能刷新日志

	// 其他级别日志也可以考虑使用异步Hook
	infolfHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.InfoLevel: infoWriter,
	}, fileFormatter)
	asyncInfoHook := NewAsyncHook(infolfHook, 1000)
	logrus.AddHook(asyncInfoHook)    // 输出到log文件夹（以异步方式）
	RegisterAsyncHook(asyncInfoHook) // 注册到全局异步Hook列表中

	warnlfHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.WarnLevel: warnWriter,
	}, fileFormatter)
	asyncWarnHook := NewAsyncHook(warnlfHook, 500)
	logrus.AddHook(asyncWarnHook)    // 输出到log文件夹（以异步方式）
	RegisterAsyncHook(asyncWarnHook) // 注册到全局异步Hook列表中

	paniclfHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.PanicLevel: panicWriter,
	}, fileFormatter)
	logrus.AddHook(paniclfHook) // panic级别保持同步处理，确保立即响应

	// 下面是另一个日志文件处理方式

	fileWriter := newDailyLumberjackLogger("all.log", rotationSize, reserveDuration, maxBackups)

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
