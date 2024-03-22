package formatter

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var LatestToken TenantAccessTokenBody // 最新的token

func FeishuRobotDetail(msg string, appName ...string) {
	go func(msg string, appName ...string) {
		_, file, line, ok := runtime.Caller(1)
		if !ok {
			log.Errorf("获取行号失败 %v,%v", file, line)
		}
		timeValue := time.Now().Format("2006-01-02 15:04:05")
		name, err := os.Hostname()
		if err != nil {
			log.Errorf("获取主机名失败 %+v", err)
		}
		if appName != nil && len(appName) > 0 {
			AppName = appName[0]
		}
		content := timeValue + "【" + AppName + "】" + name + "(" + file + ":" + strconv.Itoa(line) + "):" + msg
		feishuRobot(content)
	}(msg, appName...)
}

// 飞书机器人通知到群里 https://open.feishu.cn/document/server-docs/im-v1/message/create?appId=cli_a1cefb050579500b
func feishuRobot(textContent string) {
	// 先获取token
	if LatestToken != (TenantAccessTokenBody{}) {
		if LatestToken.RequestTime.Add(time.Duration(LatestToken.Expire) * time.Second).After(time.Now()) {
			// 可用
		} else {
			// 不可用
			LatestToken = getTenantAccessToken()
		}
	} else {
		LatestToken = getTenantAccessToken()
	}

	if LatestToken != (TenantAccessTokenBody{}) && len(LatestToken.TenantAccessToken) > 0 {
		//timeValue := time.Now().Format("2006-01-02 15:04:05")
		//content := timeValue + ",kuaima-express," + GetHostName() + ":" + textContent
		content := textContent
		client := &http.Client{}
		var data = strings.NewReader(`{
    "receive_id": "oc_0dcaa407df30d1a3415c382e397dcd0f",
    "msg_type": "text",
    "content": "{\"text\":\"` + content + `\"}"
}`)
		req, err := http.NewRequest(http.MethodPost, "https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=chat_id", data)
		if err != nil {
			log.Error(err)
		}
		req.Header.Set("Authorization", "Bearer "+LatestToken.TenantAccessToken)
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		resp, err := client.Do(req)
		if err != nil {
			log.Error(err)
		}
		if resp != nil {
			defer resp.Body.Close()
			bodyText, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Error(err)
			}
			log.Printf("飞书请求结果%s，参数 %v", bodyText, content)
		} else {
			log.Errorf("请求飞书错误， %v", textContent)
		}
	} else {
		log.Errorf("飞书机器人发送消息错误 %v", textContent)
	}
}

type TenantAccessTokenBody struct {
	Code              int `json:"code"`
	Expire            int `json:"expire"`
	RequestTime       time.Time
	Msg               string `json:"msg"`
	TenantAccessToken string `json:"tenant_access_token"`
}

func getTenantAccessToken() (tenantAccessTokenBody TenantAccessTokenBody) {
	client := &http.Client{}
	var data = strings.NewReader(`{
	"app_id": "cli_a1cefb050579500b",
	"app_secret": "PAtVnWyuRQTyRRQ1EpHQ9fnAevpYGkkV"
}`)
	req, err := http.NewRequest(http.MethodPost, "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal", data)
	if err != nil {
		log.Error(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
	}
	defer resp.Body.Close()
	if resp != nil {
		bodyText, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error(err)
		}
		err1 := json.Unmarshal(bodyText, &tenantAccessTokenBody)
		if err1 != nil {
			log.Error(err1)
		}
		tenantAccessTokenBody.RequestTime = time.Now()
	}

	return
}
