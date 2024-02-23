package formatter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
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
	SendToRobotMessage(string(payload))

	return nil
}

type MessageParamIM struct {
	TalkType   int    `form:"talk_type" json:"talk_type" binding:"required,oneof=1 2" label:"talk_type"`
	ReceiverId int    `form:"receiver_id" json:"receiver_id" binding:"required,numeric,gt=0" label:"receiver_id"`
	Text       string `form:"text" json:"text" binding:"required,max=3000" label:"text"`
	RobotId    int    `form:"robot_id" json:"robot_id" label:"robot_id"`
}

func SendToRobotMessage(msg string) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		log.Errorf("获取行号失败 %v,%v", file, line)
	}
	timeValue := time.Now().Format("2006-01-02 15:04:05")
	name, err := os.Hostname()
	if err != nil {
		log.Errorf("获取主机名失败 %+v", err)
	}
	content := timeValue + ",kuaima-express," + name + file + ":" + strconv.Itoa(line) + ":" + msg
	SendRobotMessage(content, 2, 426, 4)
}

type NewMessageParamIM struct {
	Type string `json:"type"`
	//SenderId int           `json:"sender_id"` // TODO 最好传过来，可以标记是极兔助手还是快码机器人。但是总体来说没啥关系。
	Content  string        `json:"content"`
	QuoteId  string        `json:"quote_id"`
	Mentions []interface{} `json:"mentions"`
	Receiver struct {
		ReceiverId int `json:"receiver_id"`
		TalkType   int `json:"talk_type"`
	} `json:"receiver"`
}

type NewRobotTextMessageRequest struct {
	NewMessageParamIM
	RobotId int `json:"robot_id"`
}

func SendRobotMessage(content string, talkType, ReceiverId, RobotId int) {
	defer PanicHandler()

	messageParam := NewMessageParamIM{
		Type:    "text",
		Content: content,
		Receiver: struct {
			ReceiverId int `json:"receiver_id"`
			TalkType   int `json:"talk_type"`
		}{
			ReceiverId: ReceiverId,
			TalkType:   talkType,
		},
	}

	paramMe := &NewRobotTextMessageRequest{
		NewMessageParamIM: messageParam,
		RobotId:           RobotId, // 这个可能是机器人，也可能是极兔助手（不算机器人）
	}

	marshal, err := json.Marshal(&paramMe)
	if err != nil {
		log.Error(err)
	}
	reader := strings.NewReader(string(marshal))

	headerMap := make(map[string]string)
	headerMap["KM"] = "kuaima2023"
	headerMap["content-type"] = "application/json;charset=UTF-8"

	RequestJson(http.MethodPost, "https://im.cupb.top/api/api/v1/open/talk/message/robot/text", reader, headerMap)

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
