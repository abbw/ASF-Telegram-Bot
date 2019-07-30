package main

import (
	"encoding/json"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {
	log.Println("======================================================")
	log.Println("Rakuyo的ASF-Telegram机器人 Version 0.1.1")
	log.Println("更新于2019年7月28日")
	log.Println("源码    https://github.com/rakuyo42/asf-telegram-bot")
	log.Println("示例    https://t.me/RakuyoASFBot")
	log.Println("有任何疑问请到  https://steamcn.com/t503337-1-1  反馈")
	log.Println("======================================================")
	if readConfig( "bin/config.json") {
		log.Println("配置文件已校验完毕")
		log.Println("======================================================")
		log.Println("         尝试连接到 Telegram 服务器并启动机器人")
		log.Println()
		log.Println(">>>>>>>>>>>>            请注意            <<<<<<<<<<<<")
		log.Println()
		log.Println("   如果接下来闪退大概率是因为「连接Telegram服务器失败」")
		log.Println()
		log.Println("                  请自行解决「网络问题」")
		log.Println()
		log.Println(">>>>>>>>>>>>          少女折寿中          <<<<<<<<<<<<")
		startBot()
	}
	log.Println("将在1分钟后退出...")
	time.Sleep(60 * 1000000000)
}

func startBot() {
	bot, err := tgbotapi.NewBotAPI(CONFIG.BotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Println("连接到机器人成功")
	log.Println("======================================================")
	log.Printf("已获得机器人 %s 的控制权", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	log.Printf("%s正在辛勤工作中...", bot.Self.UserName)
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		var reply string
		if update.Message.Chat.ID != CONFIG.ChatID {
			speakName := update.Message.Chat.FirstName
			if update.Message.Chat.LastName != "" {
				speakName += " " + update.Message.Chat.LastName
			}
			reply = speakName + "，你没有此bot的控制权。\n你的Username：" + update.Message.Chat.UserName +
				"\n你的Chat ID：" + strconv.FormatInt(update.Message.Chat.ID, 10)
		} else {
			var query_str string
			var CommandPrefix string = "!"
			if strings.HasPrefix(update.Message.Text, "/") {
				query_str = strings.Replace(update.Message.Text, "/", CommandPrefix, 1)
			} else if strings.HasPrefix(update.Message.Text, CommandPrefix) {
				query_str = update.Message.Text
			} else {
				continue
			}
			reply = queryASF(query_str)
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
	}
}

func queryASF(query string) string {

	/* 构造URL */
	url := CONFIG.IPCUrl
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	url += "Api/Command/" + query

	/* 构造请求 */
	req, _ := http.NewRequest("POST", url, nil)

	/* 添加身份认证请求头 */
	if CONFIG.IPCPassword != "" {
		req.Header.Set("Authentication", CONFIG.IPCPassword)
	}

	/* 发送请求并获取响应 */
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return "0xF1"
	}

	/* 解析响应 */
	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	var queryResult string = resp.Status
	if err := json.Unmarshal([]byte(body), &result); err == nil {
		if queryResult == "200 OK" {
			queryResult = result["Result"].(string)
		}
	}
	return queryResult
}

func readConfig(config_file string) bool {
	/* 取得当前工作目录 */
	/* 另一种写法
	root_path, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Println(err)
	}
	*/
	root_path, _ := os.Getwd()
	file_path := filepath.Join(root_path, config_file)
	log.Printf("读取配置文件 %s", file_path)

	/* 如果不存在配置文件则创建 */
	_, err := os.Stat(file_path)
	if os.IsNotExist(err) {
		file, err := os.Create(file_path)
		if err != nil {
			log.Println(err)
		}
		defer file.Close()
	}

	/* 读取配置 */
	configdata, err := ioutil.ReadFile(file_path)
	if err := json.Unmarshal(configdata, &CONFIG); err != nil {
		/* 读取失败则初始化为缺省值 */
		CONFIG.IPCUrl = "127.0.0.1:1242"
		if configsdata, err := json.MarshalIndent(CONFIG, "", "\t"); err != nil {
			log.Panic(err)
		} else {
			if err = ioutil.WriteFile(file_path, configsdata, 0666); err != nil {
				log.Panic(err)
			}
		}
		log.Printf("已生成配置文件config.json")
		log.Printf("请前往%s编辑后重启本程序", file_path)
		return false
	} else {
		return testIPC()
	}
}

func testIPC() bool {
	if CONFIG.BotToken == "" {
		log.Printf("缺少bot token，将无法连接到机器人。")
	}
	if CONFIG.ChatID == 0 {
		log.Printf("缺少chat id，将无法识别你的telegram账号。")
	} else {
		log.Printf("尝试连接预设的 ASF-IPC(%s) ...", CONFIG.IPCUrl)
		ret := queryASF("version")
		if strings.Contains(ret, "ASF") {
			log.Printf("连接到 ASF-IPC(%s) 成功", CONFIG.IPCUrl)
			return true
		} else if strings.HasPrefix(ret, "0xF1") {
			log.Printf("连接失败")
		} else if strings.HasPrefix(ret, "400") {
			log.Printf("请求失败 返回的响应为%s", ret)
		} else if strings.HasPrefix(ret, "401") {
			log.Printf("请求的ASF-IPC设置了IPCPassword 请检查本地设置的IPC密码是否正确")
		} else if strings.HasPrefix(ret, "403") {
			log.Printf("请求的ASF-IPC设置了IPCPassword 失败多次已被暂时封禁 请1小时后再试")
		} else if strings.HasPrefix(ret, "500") {
			log.Printf("ASF在服务请求时遇到意外错误 请检查ASF日志")
		} else if strings.HasPrefix(ret, "503") {
			log.Printf("ASF在请求第三方资源时遇到错误 请稍后再试")
		} else {
			log.Printf("未知错误! 连接 ASF-IPC(%s) 失败")
		}
	}
	return false
}

var CONFIG ConfigStruct

type ConfigStruct struct {
	BotToken    string `json:"tg_bot_token"`
	ChatID      int64  `json:"tg_chat_id"`
	IPCUrl      string `json:"asf_ipc_url"`
	IPCPassword string `json:"asf_ipc_password"`
}
