package main

import (
	"fmt"

	formatter "github.com/aohanhongzhi/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("--- GORM 错误邮件报警示例 ---\n")

	// 初始化日志
	formatter.LogInitWithLevel(false, "gorm-example", log.DebugLevel)

	// 可选：启用GORM错误邮件报警（需要有效的邮件服务器配置）
	// 注意：这里使用虚拟配置，仅作为示例

	err := formatter.EnableGormErrorMail(
		"gorm-example",
		"smtp.qq.com",
		587,
		"your-email@qq.com",
		"alert-email@qq.com",
		"your-email@qq.com",
		"your-password", // QQ邮箱的授权码
	)
	if err != nil {
		log.Errorf("启用GORM错误邮件报警失败: %v", err)
	} else {
		log.Info("GORM错误邮件报警已启用")
	}

	// 模拟GORM错误日志（实际使用中这些会由gormv2-logrus自动生成）
	log.WithFields(log.Fields{
		"component": "gorm",
	}).Error("[Error 1406 (22001): Data too long for column 'network_info' at row 1] UPDATE `npc_user_fee_model` SET `fee`=3600,`network_info`=[...] WHERE `ID` = 527")

	// 普通错误日志（不会触发邮件报警）
	log.Error("这是一条普通的错误日志")

	// 检查GORM邮件报警状态
	enabled := formatter.IsGormErrorMailEnabled()
	fmt.Printf("GORM错误邮件报警启用状态: %v\n", enabled)

	fmt.Println("\n示例完成。实际使用时，请配置有效的邮件服务器信息。")
}
