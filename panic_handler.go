package formatter

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"time"
)

// https://blog.csdn.net/xia_xing/article/details/80597472
// 异常处理
func PanicHandler() {
	errs := recover()
	if errs == nil {
		// 没异常发生
		return
	} else {
		_, file, line, ok := runtime.Caller(1)
		if !ok {
			log.Errorf("获取行号失败 %v,%v", file, line)
		} //https://github.com/sirupsen/logrus/issues/181
		exeName := os.Args[0] //获取程序名称

		now := time.Now()  //获取当前时间
		pid := os.Getpid() //获取进程ID

		time_str := now.Format("20060102150405")                             //设定时间格式
		fname := fmt.Sprintf("%s-pid%d-%s-dump.log", exeName, pid, time_str) //保存错误信息文件名:程序名-进程ID-当前时间（年月日时分秒）
		fmt.Println("dump to file ", fname)

		f, err := os.Create(fname)
		if err != nil {
			return
		}
		defer f.Close()

		errInfo := fmt.Sprintf("%v\r\n", errs)
		f.WriteString(errInfo) //输出panic信息
		f.WriteString("========\r\n")

		stackInfo := string(debug.Stack())
		f.WriteString(stackInfo) //输出堆栈信息

		// 发生错误
		content := "服务器记录文件名:" + fname + "\n" + stackInfo
		log.WithField("panic", errs).Error("we panicked!发生地方" + file + ":" + strconv.Itoa(line) + "\n" + content)

	}
}
