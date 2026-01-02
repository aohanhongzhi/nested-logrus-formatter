package notification

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

const (
	gormErrorFormat = "20060102 15:04:05"
)

// GormErrorMailHook 专门用于GORM错误日志的邮件报警Hook
type GormErrorMailHook struct {
	AppName  string
	Host     string
	Port     int
	From     *mail.Address
	To       *mail.Address
	Username string
	Password string
	Enable   bool // 是否启用GORM错误邮件报警
}

// 创建GORM错误邮件报警Hook
func NewGormErrorMailHook(appName, host string, port int, from, to, username, password string) (*GormErrorMailHook, error) {
	// 验证服务器连接
	conn, err := net.DialTimeout("tcp", host+":"+strconv.Itoa(port), 3*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// 验证邮件地址
	sender, err := mail.ParseAddress(from)
	if err != nil {
		return nil, err
	}
	receiver, err := mail.ParseAddress(to)
	if err != nil {
		return nil, err
	}

	return &GormErrorMailHook{
		AppName:  appName,
		Host:     host,
		Port:     port,
		From:     sender,
		To:       receiver,
		Username: username,
		Password: password,
		Enable:   true,
	}, nil
}

// Levels 返回处理的日志级别
func (hook *GormErrorMailHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}

// Fire 处理日志条目
func (hook *GormErrorMailHook) Fire(entry *logrus.Entry) error {
	if !hook.Enable {
		return nil
	}

	// 检测是否为GORM错误日志
	if !isGormErrorLog(entry) {
		return nil
	}

	// 发送邮件
	return hook.sendGormErrorMail(entry)
}

// isGormErrorLog 检测是否为GORM错误日志
func isGormErrorLog(entry *logrus.Entry) bool {

	for key, value := range entry.Data {
		if key == "error" {
			if _, ok := value.(*mysql.MySQLError); ok {
				return true
			}
		}
	}

	// 检查Message是否包含GORM错误特征
	if strings.Contains(entry.Message, "[Error ") && strings.Contains(entry.Message, ":") {
		// 检查是否包含gormv2-logrus的调用栈特征
		if entry.HasCaller() && strings.Contains(entry.Caller.File, "gormv2-logrus") {
			return true
		}
		// 检查数据字段中是否有gorm相关信息
		for key, value := range entry.Data {
			if key == "component" && strings.Contains(fmt.Sprintf("%v", value), "gorm") {
				return true
			}
		}
	}
	return false
}

// sendGormErrorMail 发送GORM错误邮件
func (hook *GormErrorMailHook) sendGormErrorMail(entry *logrus.Entry) error {
	auth := smtp.PlainAuth("", hook.Username, hook.Password, hook.Host)

	// 创建 TLS 连接
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // 如果需要跳过证书验证，可以设置为 true
		ServerName:         hook.Host,
	}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", hook.Host, hook.Port), tlsConfig)
	if err != nil {
		return err
	}

	message := hook.createGormErrorMailMessage(entry)

	client, err := smtp.NewClient(conn, hook.Host)
	if err != nil {
		_ = conn.Close()
		return err
	}

	// 确保关闭
	defer func() {
		if client != nil {
			if e := client.Quit(); e != nil {
				if !strings.Contains(e.Error(), "OK") {
				}
			}
		}
		_ = conn.Close()
	}()

	// 使用 AUTH 进行身份验证
	if err := client.Auth(auth); err != nil {
		if strings.Contains(err.Error(), "535") {
			return err
		} else {
			return err
		}
	}

	// 设置发件人和收件人
	var toPerson string

	if err := client.Mail(hook.From.Address); err != nil {
		return err
	}
	for _, addr := range []string{hook.To.Address} {
		//for _, addr := range []string{"3227556776@qq.com", "3028536139@qq.com", "18175306923@163.com"} {
		if toPerson == "" {
			toPerson = addr
		} else {
			toPerson = toPerson + "," + addr
		}
		if err := client.Rcpt(addr); err != nil {
			return err
		}
	}

	// 发送邮件内容
	wc, err := client.Data()
	if err != nil {
		return err
	}
	if wc != nil {
		defer wc.Close()

		_, err = fmt.Fprintf(wc, "%v", message)
		if err != nil {
			return err
		}
	}

	return nil
}

// createGormErrorMailMessage 创建GORM错误邮件消息
func (hook *GormErrorMailHook) createGormErrorMailMessage(entry *logrus.Entry) *bytes.Buffer {
	body := fmt.Sprintf("[%s] GORM Database Error 数据库错误\n\n时间: %s\n错误信息: %s",
		hook.AppName,
		entry.Time.Format("2006-01-02 15:04:05"),
		entry.Message)

	subject := fmt.Sprintf("[%s] GORM Database Error 数据库错误报警", hook.AppName)

	// 添加调用栈信息
	if entry.HasCaller() {
		body += fmt.Sprintf("\n\n调用位置:\n文件: %s\n行号: %d\n函数: %s",
			entry.Caller.File,
			entry.Caller.Line,
			entry.Caller.Function)
	}

	// 添加额外字段信息
	if len(entry.Data) > 0 {
		body += "\n\n额外信息:\n"
		for key, value := range entry.Data {
			body += fmt.Sprintf("%s: %v\n", key, value)
		}
	}

	contents := fmt.Sprintf("From: %v\nTo: %v\nSubject: %s\nContent-Type: text/plain; charset=UTF-8\n\n%s",
		hook.From.Address, hook.To.Address, subject, body)

	message := bytes.NewBufferString(contents)
	return message
}

// 启用GORM错误邮件报警（全局函数，方便调用）
var globalGormErrorMailHook *GormErrorMailHook

func EnableGormErrorMail(appName, host string, port int, from, to, username, password string) error {
	hook, err := NewGormErrorMailHook(appName, host, port, from, to, username, password)
	if err != nil {
		return fmt.Errorf("创建GORM错误邮件Hook失败: %v", err)
	}

	globalGormErrorMailHook = hook
	logrus.AddHook(hook)
	return nil
}

// DisableGormErrorMail 禁用GORM错误邮件报警
func DisableGormErrorMail() {
	if globalGormErrorMailHook != nil {
		globalGormErrorMailHook.Enable = false
	}
}

// IsGormErrorMailEnabled 检查GORM错误邮件报警是否启用
func IsGormErrorMailEnabled() bool {
	return globalGormErrorMailHook != nil && globalGormErrorMailHook.Enable
}

// GetGormErrorMailHook 获取全局GORM错误邮件报警Hook实例
func GetGormErrorMailHook() *GormErrorMailHook {
	return globalGormErrorMailHook
}
