package formatter

import (
	"path/filepath"
	"sync"
	"unicode"

	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// TagRouterHook 根据 entry.Data["tag"] 将日志额外写入 tag 专属文件
// 注意：该 Hook 不改变原有日志流，只做“旁路”追加
type TagRouterHook struct{}

func (h *TagRouterHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

var tagHookCache sync.Map // map[string]logrus.Hook

func sanitizeTagFilename(s string) string {
	if s == "" {
		return "unknown"
	}
	runes := []rune(s)
	for i, r := range runes {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.' {
			continue
		}
		runes[i] = '_'
	}
	return string(runes)
}

func getOrCreateTagHook(tagName string) logrus.Hook {
	if h, ok := tagHookCache.Load(tagName); ok {
		return h.(logrus.Hook)
	}
	baseDir := LogBaseDir
	if baseDir == "" {
		baseDir = "."
	}
	filePath := filepath.Join(baseDir, "/log/tag/", tagName+".log")
	MkLogdir(filepath.Dir(filePath))

	writer := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    int(GlobalRotationSize) / (1024 * 1024),
		MaxBackups: GlobalMaxBackups,
		MaxAge:     int(GlobalReserveDuration.Hours() / 24),
		Compress:   true,
	}

	fileFormatter := &Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		NoColors:        false,
		HideKeys:        true,
		NoFieldsSpace:   false,
		FieldsOrder:     []string{"component", "category", "req"},
	}

	writerMap := lfshook.WriterMap{
		logrus.TraceLevel: writer,
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.PanicLevel: writer,
		logrus.FatalLevel: writer,
	}

	lfHook := lfshook.NewHook(writerMap, fileFormatter)
	tagHookCache.Store(tagName, lfHook)
	return lfHook
}

func (h *TagRouterHook) Fire(entry *logrus.Entry) error {
	val, ok := entry.Data["tag"]
	// 未带 tag 时写入 no-tag.log
	if !ok {
		lfHook := getOrCreateTagHook("no-tag")
		return lfHook.Fire(entry)
	}
	// 转字符串并安全化文件名
	tagName := ""
	switch v := val.(type) {
	case string:
		tagName = v
	default:
		tagName = "unknown"
	}
	if tagName == "" {
		tagName = "no-tag"
	}
	tagName = sanitizeTagFilename(tagName)

	// 获取或创建对应的 hook 并转发
	lfHook := getOrCreateTagHook(tagName)
	return lfHook.Fire(entry)
}

// EnableTagFileRouter 启用 Tag 旁路写入能力
// 默认查找字段名 "tag"，例如： log.WithField("tag", "order").Info("created")
// 开闭原则：不改变任何现有接口，由业务层在初始化后主动调用启用
func EnableTagFileRouter() {
	// 包装为异步，避免阻塞主流程
	async := NewAsyncHook(&TagRouterHook{}, 500)
	logrus.AddHook(async)
	RegisterAsyncHook(async)
}
