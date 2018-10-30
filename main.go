// emailTools project main.go
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/qiniu/iconv"

	"github.com/larspensjo/config"

	"github.com/smartwalle/going/email"
)

var TOPIC = make(map[string]string)
var sysdate = time.Now().Format("2006-01")
var ti = time.Now().Format("20060102")
var logti = time.Now().Format("2006/01/02 03:04:05 PM")

func main() {
	readLogin()
	SendEmail()
}

func SendEmail() {
	os.IsExist(os.Mkdir("log", os.ModePerm))
	logFile, _ := os.OpenFile("log/"+sysdate+".txt", os.O_RDWR|os.O_CREATE, 0666)
	SEEK_END, _ := logFile.Seek(0, os.SEEK_END)
	//ExportFileName := generate(logFile)
	defer logFile.Close()
	_, _ = logFile.WriteAt([]byte("\r\n"), SEEK_END)
	logFile.WriteString(logti + "\r\n")

	cd, err := iconv.Open("utf-8", "gbk")
	erro(err)
	defer cd.Close()
	signature, _ := ioutil.ReadFile(TOPIC["signature"])

	ToEmailList := TOPIC["ToEmailist"]
	CcEmailList := TOPIC["CcEmailist"]
	//fmt.Println("hello1")
	var config = &email.MailConfig{}
	config.Username = TOPIC["username"]
	config.Host = TOPIC["host"]
	config.Password = TOPIC["password"]
	config.Port = TOPIC["port"]
	config.Secure = false

	var eContent string
	title := TOPIC["emailTitle"]
	title = cd.ConvString(title)
	var e = email.NewTextMessage(title, "")
	if TOPIC["attach"] == "no" {
		logFile.WriteString("This email is not attached\r\n")
	} else {
		attachNamePrefix := cd.ConvString(TOPIC["attachNamePrefix"])
		attachName := attachNamePrefix + ti + TOPIC["attachNameStffix"]
		attachNameLog := TOPIC["attachNamePrefix"] + ti + TOPIC["attachNameStffix"]
		attachExist := false
		attachPath := TOPIC["attachPath"] + "\\" + sysdate
		files, _ := ListDir(attachPath, "")
		for _, v := range files {
			if strings.Contains(v, attachName) {
				attachExist = true
			}
		}
		if attachExist {
			logFile.WriteString("Generate the attachment in/" + TOPIC["attachP"] + sysdate + "/" + attachNameLog + "\r\n")
			e.AttachFile(TOPIC["attachP"] + sysdate + "/" + attachName)
		} else {
			logFile.WriteString("Attachment not found.\r\nPlease check this directory exists:" + TOPIC["attachP"] + attachNameLog + "\r\n")
			eContent = "The email attachment is not found, please check the reason in time!!!\r\nThe attachment theory exists in:" + TOPIC["attachP"] + sysdate + "/" + attachName + "\r\n"
		}
	}
	e.From = TOPIC["from"]
	//get current dealer's email address
	//e.To = []string{ToEmailList[0]}
	if ToEmailList != "" {
		e.To = strings.Split(ToEmailList, ",")
		logFile.WriteString("has sent To " + ToEmailList + "\r\n")
	} else {
		logFile.WriteString("No email address.\r\n")
	}

	//fmt.Println(ToEmailList[0])
	//get current dealer's cc email address
	if len(TOPIC["cc"]) != 0 {
		//e.Cc = []string{CcEmailList[0]}
		tmpemail := TOPIC["cc"]
		if CcEmailList != "" {
			tmpemail = tmpemail + "," + CcEmailList
		}
		e.Cc = strings.Split(tmpemail, ",")
		logFile.WriteString("has sent Cc  :" + tmpemail + "\r\n")
	} else {
		logFile.WriteString("Please set up the email box that I want to cc in the configuration file.\r\n")
	}

	//	e.Cc = []string{SMEmailList[0]}
	//e.Bcc = []string{"dzamd@dongzhengafc.com"}
	//e.Bcc = []string{TOPIC["bcct"]}
	emailBody := TOPIC["BodyPrefix"] + ti + TOPIC["BodyStuffix"]
	b, _ := ioutil.ReadFile(emailBody)
	if TOPIC["body"] != "no" {
		if string(b) == "" {
			eContent = eContent + "未检测到邮件内容，或者内容为空，请检查是否已生成！！\r\n邮件内容理论存在于：" + emailBody + "\r\n"
			logFile.WriteString("The contents of the message are not detected, or the content is empty, please check if it has been generated!\r\nThe content theory of the email exists in:" + emailBody + "\r\n")
			e.Content = eContent + string(signature)

		} else {
			e.Content = eContent + string(b) + string(signature)
		}
	} else {
		e.Content = eContent + string(signature)
		logFile.WriteString("Don't read the email\r\n")
	}

	logFile.WriteString("-----------------------------\r\n")
	//get current dealer's email attachment
	//e.AttachFile(ExportFileName)
	err = email.SendMail(config, e)
	if err != nil {
		logFile.WriteString(err.Error())
	}

	erro(err)
}

func readLogin() {
	var (
		configFile = flag.String("configfile", "config.ini", "General configuration file")
	)
	flag.Parse()
	cfg, err := config.ReadDefault(*configFile)
	if err != nil {
		fmt.Println("read ini error")
		return
	}
	if cfg.HasSection("exe") {
		section, err := cfg.SectionOptions("exe")
		if err == nil {
			for _, v := range section {
				options, err := cfg.String("exe", v)
				if err == nil {

					TOPIC[v] = options
				}
			}
		}
	}
}

func ListDir(dirPth string, suffix string) (files []string, err error) {
	files = make([]string, 0, 10) //初始化file切片，预留十个位置
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}
	PthSep := string(os.PathSeparator)
	suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写
	for _, fi := range dir {
		if fi.IsDir() { // 忽略目录
			continue
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) { //匹配文件
			files = append(files, dirPth+PthSep+fi.Name())
		}
	}
	return files, nil
}
func erro(err error) {
	if err != nil {
		fmt.Println("出错了", err)
	}
}
