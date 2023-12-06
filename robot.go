package formatter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type RobotLog struct {
	AppName string
}

func NewRobotLogger(AppName string) *RobotLog {
	return &RobotLog{AppName}
}

func (hook *RobotLog) Levels() []log.Level {
	return []log.Level{
		log.PanicLevel,
		log.FatalLevel,
		log.ErrorLevel,
	}
}

func (hook *RobotLog) Fire(entry *log.Entry) error {
	data := make(log.Fields)
	for k, v := range entry.Data {
		data[k] = v
	}
	data["app"] = hook.AppName
	if entry.HasCaller() {
		fileVal := fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)
		data["location"] = fileVal
	}
	data["message"] = entry.Message
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	SendRobotNoticeGroup(string(payload))

	return nil
}

type MessageParamIM struct {
	TalkType   int    `form:"talk_type" json:"talk_type" binding:"required,oneof=1 2" label:"talk_type"`
	ReceiverId int    `form:"receiver_id" json:"receiver_id" binding:"required,numeric,gt=0" label:"receiver_id"`
	Text       string `form:"text" json:"text" binding:"required,max=3000" label:"text"`
	RobotId    int    `form:"robot_id" json:"robot_id" label:"robot_id"`
}

func SendRobotMessage(content string, talkType, ReceiverId, RobotId int) {
	req := &MessageParamIM{
		TalkType:   talkType, // 私聊
		Text:       content,
		ReceiverId: ReceiverId, // 同一发送给机器人助手
		RobotId:    RobotId,    // 可以给 token信息即可识别身份
	}

	marshal, err := json.Marshal(&req)
	if err != nil {
		log.Error(err)
	}
	reader := strings.NewReader(string(marshal))

	headerMap := make(map[string]string)
	headerMap["KM"] = "kuaima2023"
	headerMap["content-type"] = "application/json;charset=UTF-8"
	RequestJson("POST", "https://im.cupb.top/api/api/v1/open/talk/message/robot/text", reader, headerMap)
}

func SendRobotNoticeGroup(content string) {
	timeValue := time.Now().Format("2006-01-02 15:04:05")
	name, err := os.Hostname()
	if err != nil {
		log.Errorf("获取主机名失败 %+v", err)
	}
	content = timeValue + "," + name + ":" + content
	SendRobotMessage(content, 2, 426, 4)
}

var jtClient = &http.Client{}

func RequestJson(method string, url string, paramBody io.Reader, headerMap map[string]string) {
	req, err := http.NewRequest(method, url, paramBody)
	if err != nil {
		log.Error(err)
		return
	}

	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")

	if len(headerMap) == 0 {
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	} else {
		for headerKey, headerValue := range headerMap {
			req.Header.Set(headerKey, headerValue)
		}
	}

	_, err = jtClient.Do(req)
	if err != nil {
		log.Error(err)
		return
	}

	return
}
