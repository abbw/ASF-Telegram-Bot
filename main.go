package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/net/proxy"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("<Panic>", err)
		}
		//fmt.Println("将在1分钟后退出...")
		time.Sleep(10 * 1000000000)
		//fmt.Println("按任意键退出...")
	}()

	printLog("beginning")
	if applyConfig("config.json") {
		fmt.Println("读取配置文件成功!")
	} else {
		fmt.Println("已初始化配置文件!")
	}
	if CONFIG.BotToken == "" {
		fmt.Print("请输入bot token: ")
		fmt.Scanf("%s", &CONFIG.BotToken)
	}
	printLog("startbot")
	startBot(getProxyClient())
}

func startBot(client *http.Client) {
	bot, err := tgbotapi.NewBotAPIWithClient(CONFIG.BotToken, client)
	if err != nil {
		fmt.Println("连接到机器人失败!")
		panic(err)
	} else {
		saveConfig()
		fmt.Println("成功连接到机器人!")
		fmt.Printf("已获得机器人 [%s](@%s) 的控制权!\n", bot.Self.FirstName, bot.Self.UserName)
	}
	TempVerification := getRandString(10)
	if CONFIG.ChatID == 0 {
		var master_chat_id int64
		fmt.Println("开始设置机器人归属的Telegram账号...")
		fmt.Println("请输入你的Telegram账号的Chat id: (跳过请直接回车)")
		fmt.Println("(跳过后可直接与机器人对话设置或者编辑配置文件config.json来设置)")
		fmt.Scanln(&master_chat_id)
		if master_chat_id > 0 {
			CONFIG.ChatID = master_chat_id
			saveConfig()
		} else {
			fmt.Printf("请尽快私聊机器人[%s](@%s)消息 %s 以验证你的telegram账号!\n", bot.Self.FirstName, bot.Self.UserName, TempVerification)
		}
	}
	var isASFReady bool
	var nextSet bool
	fmt.Println("开始测试ASF-IPC连接的畅通性...")
	if test_ASF_IPC() {
		fmt.Printf("ASF-IPC正在正常工作!\n")
		isASFReady = true
	} else {
		fmt.Printf("测试ASF-IPC连接未通过，请检查配置是否正确!\n")
		fmt.Printf("编辑config.json或与机器人对话来重新配置!\n")
		isASFReady = false
	}
	fmt.Println("======================================================")

	if CONFIG.DebugBot {
		bot.Debug = true
	} else {
		bot.Debug = false
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)

	fmt.Printf("%s(https://t.me/%s) 正在辛勤工作中...\n", bot.Self.FirstName, bot.Self.UserName)
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		fmt.Printf("[%s]: %s\n", update.Message.From.UserName, update.Message.Text)

		var reply string
		if CONFIG.ChatID == 0 {
			if len(update.Message.Text) == 10 {
				if update.Message.Text != TempVerification {
					reply = "验证码错误！\n请重新输入："
				} else {
					master := update.Message.Chat
					reply = "验证成功！\n已确认 " + master.FirstName + " " + master.LastName + "(@" +
						master.UserName + " " + strconv.FormatInt(master.ID, 10) + ") 为机器人所有者！"
					CONFIG.ChatID = master.ID
					saveConfig()
				}
			} else {
				reply = "还未设置机器人归属！\n请输入服务端生成的临时验证码："
			}
		} else if update.Message.Chat.ID != CONFIG.ChatID {
			speakName := update.Message.Chat.FirstName
			if update.Message.Chat.LastName != "" {
				speakName += " " + update.Message.Chat.LastName
			}
			reply = speakName + "，你没有此bot的控制权。\n你的Username: @" + update.Message.Chat.UserName +
				"\n你的Chat ID: " + strconv.FormatInt(update.Message.Chat.ID, 10)
		} else if !isASFReady && !nextSet {
			reply = "ASF尚未准备好，请重新配置。\n请输入ASF-IPC的URL及密码(用空格隔开)："
			nextSet = true
		} else if nextSet {
			ipc_config := strings.SplitN(update.Message.Text, " ", 2)
			if len(ipc_config) > 1 {
				CONFIG.IPCUrl = ipc_config[0]
				CONFIG.IPCPassword = ipc_config[1]
				if test_ASF_IPC() {
					saveConfig()
					reply = "测试通过！可以正常使用机器人。"
					isASFReady = true
					nextSet = false
				} else {
					reply = "连接IPC失败，请重新设置。"
				}
			} else {
				reply = "格式错误！请输入 [IPC地址 IPC密码] 的格式。如\n127.0.0.1:1242 password"
			}
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
		fmt.Printf("[%s]: %s\n", bot.Self.FirstName, reply)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)
	}
}

func printLog(keyword string) {
	if keyword == "beginning" {
		fmt.Println("======================================================")
		fmt.Println("Rakuyo的ASF-Telegram机器人 Version 0.9.4")
		fmt.Println("最后更新于2019年10月31日")
		fmt.Println("源码    https://github.com/rakuyo42/ASF-Telegram-Bot")
		fmt.Println("示例    https://t.me/RakuyoASFBot")
		fmt.Println("有任何疑问请到  https://keylol.com/t503337-1-1  反馈")
		fmt.Println("======================================================")
	} else if keyword == "startbot" {
		fmt.Println("======================================================")
		fmt.Println("        尝试连接到 Telegram 服务器并启动机器人")
		fmt.Println()
		fmt.Println(">>>>>>>>>>>>            请注意            <<<<<<<<<<<<")
		fmt.Println()
		fmt.Println("                         出现")
		fmt.Println("    <Panic> Post https://api.telegram.org/bot......")
		fmt.Println("            则为「连接Telegram服务器失败」")
		fmt.Println("                请自行解决「网络问题」")
		fmt.Println("  (比如编辑配置文件config.json设置http或者socket代理)")
		fmt.Println("    (最便捷&一劳永逸的方法是在国外vps上运行本程序)")
		fmt.Println()
		fmt.Println(">>>>>>>>>>>>          少女折寿中          <<<<<<<<<<<<")
	}
}

func gen_API_URL(query string, api string) string {
	url := CONFIG.IPCUrl
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	if api == "asf" {
		url += "Api/ASF/" + query
	} else if api == "bot" {
		url += "Api/Bot/" + query
	} else if api == "command" {
		url += "Api/Command/"
	}
	return url
}

func json_to_string(resp_body []byte, resp_status string) string {
	resp_msg := resp_status
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp_body), &result); err == nil && resp_status == "200 OK" {
		resp_msg = result["Result"].(string)
	}
	return resp_msg
}

func queryASF(command string) string {
	/* 构造请求 */
	req_method := "POST"
	req_url := gen_API_URL(command, "command")
	req_body := bytes.NewBuffer([]byte("{\"Command\": \"" + command + "\"}"))
	fmt.Printf("携带json %s 访问 %s\n", req_body, req_url)
	req, _ := http.NewRequest(req_method, req_url, req_body)
	req.Header.Set("Content-Type", "application/json")
	/* 添加身份认证请求头 */
	if CONFIG.IPCPassword != "" {
		req.Header.Set("Authentication", CONFIG.IPCPassword)
	}
	/* 发送请求并解析响应 */
	client := &http.Client{}
	if resp, err := client.Do(req); err != nil {
		fmt.Println(err)
		return "0xF2"
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		return json_to_string(body, resp.Status)
	}
}

func test_ASF_IPC() bool {
	fmt.Println("======================================================")
	fmt.Printf("尝试连接预设的 ASF-IPC(%s) ...\n", CONFIG.IPCUrl)
	req, _ := http.NewRequest("GET", gen_API_URL("", "asf"), nil)
	if CONFIG.IPCPassword != "" {
		req.Header.Set("Authentication", CONFIG.IPCPassword)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	var ret string
	if err != nil {
		fmt.Println(err)
		ret = "0xF1"
	} else {
		ret = resp.Status
	}
	if strings.HasPrefix(ret, "200") {
		fmt.Printf("连接到 ASF-IPC(%s) 成功!\n", CONFIG.IPCUrl)
		return true
	} else if strings.HasPrefix(ret, "0xF1") {
		fmt.Printf("连接失败!\n")
	} else if strings.HasPrefix(ret, "400") {
		fmt.Printf("请求失败! 返回的响应为 %s !\n", ret)
	} else if strings.HasPrefix(ret, "401") {
		fmt.Printf("请求的ASF-IPC设置了IPCPassword! 请检查本地设置的IPC密码是否正确!\n")
	} else if strings.HasPrefix(ret, "403") {
		fmt.Printf("请求的ASF-IPC设置了IPCPassword! 失败多次已被暂时封禁! 请1小时后再试!\n")
	} else if strings.HasPrefix(ret, "500") {
		fmt.Printf("ASF在服务请求时遇到意外错误! 请检查ASF日志!\n")
	} else if strings.HasPrefix(ret, "503") {
		fmt.Printf("ASF在请求第三方资源时遇到错误! 请稍后再试!\n")
	} else {
		fmt.Printf("未知错误! 连接 ASF-IPC(%s) 失败!\n")
	}
	return false
}

func getProxyClient() *http.Client {
	var proxy_url string
	if CONFIG.SocketProxy != "" {
		proxy_url = strings.ToLower(CONFIG.SocketProxy)
		if len(proxy_url) < 6 {
			proxy_url = "127.0.0.1:" + proxy_url
		}
		dialer, err := proxy.SOCKS5("tcp", proxy_url, nil, proxy.Direct)
		if err != nil {
			fmt.Fprintln(os.Stderr, "can't connect to the proxy:", err)
			os.Exit(1)
		}
		httpTransport := &http.Transport{}
		httpClientWithProxy := &http.Client{Transport: httpTransport}
		httpTransport.Dial = dialer.Dial
		fmt.Printf("已设置Socket代理，优先使用socket，尝试使用Socket5代理%s\n", proxy_url)
		return httpClientWithProxy
	}
	if CONFIG.HTTPProxy != "" {
		proxy_url = strings.ToLower(CONFIG.HTTPProxy)
		if strings.HasPrefix(proxy_url, "https://") {
			os.Setenv("HTTPS_PROXY", proxy_url)
		} else {
			if len(proxy_url) < 6 {
				proxy_url = "http://127.0.0.1:" + proxy_url
			} else if !strings.HasPrefix(proxy_url, "http://") {
				proxy_url = "http://" + proxy_url
			}
			os.Setenv("HTTP_PROXY", proxy_url)
		}
		fmt.Printf("已设置HTTP代理，尝试使用HTTP代理%s\n", proxy_url)
	} else {
		fmt.Printf("没有设置代理，尝试直连https://api.telegram.org/\n")
	}
	return &http.Client{}
}

func getFile(file_name string) (string, error) {
	/* 取得当前工作目录 */
	/* 另一种写法
	root_path, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println(err)
	}
	*/
	root_path, _ := os.Getwd()
	file_path := filepath.Join(root_path, file_name)
	//fmt.Printf("尝试读取配置文件 %s ...\n", file_path)

	/* 如果不存在配置文件则创建 */
	_, err := os.Stat(file_path)
	if os.IsNotExist(err) {
		fmt.Println("未检测到已存在的配置文件!")
		file, err := os.Create(file_path)
		if err != nil {
			fmt.Println(err)
		}
		defer file.Close()
	}
	return file_path, nil
}

func applyConfig(config_file string) bool {
	fmt.Print("尝试读取配置文件... ")
	file_path, err := getFile(config_file)
	if err != nil {
		fmt.Println(err)
	}

	/* 读取配置 */
	configdata, err := ioutil.ReadFile(file_path)
	if err := json.Unmarshal(configdata, &CONFIG); err != nil {
		/* 读取失败则初始化为缺省值 */
		fmt.Print("尝试初始化配置文件... ")
		CONFIG.IPCUrl = "127.0.0.1:1242"
		if configsdata, err := json.MarshalIndent(CONFIG, "", "\t"); err != nil {
			fmt.Println(err)
		} else {
			if err = ioutil.WriteFile(file_path, configsdata, 0666); err != nil {
				fmt.Println(err)
			}
		}
		return false
	} else {
		return true
	}
}

func saveConfig() {
	configsdata, err := json.MarshalIndent(CONFIG, "", "\t")
	if err != nil {
		fmt.Println(err)
	}
	file, err := getFile("config.json")
	if err != nil {
		fmt.Println(err)
	}
	err = ioutil.WriteFile(file, configsdata, 0666)
	if err != nil {
		fmt.Println(err)
	}
}

func getRandString(len int) string {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		var b int
		str_type := rand.Intn(3)
		if str_type == 0 {
			b = r.Intn(10) + 48
		} else if str_type == 1 {
			b = r.Intn(26) + 65
		} else {
			b = r.Intn(26) + 97
		}
		bytes[i] = byte(b)
	}
	return string(bytes)
}

var CONFIG ConfigStruct

type ConfigStruct struct {
	BotToken    string `json:"tg_bot_token"`
	ChatID      int64  `json:"tg_chat_id"`
	IPCUrl      string `json:"asf_ipc_url"`
	IPCPassword string `json:"asf_ipc_password"`
	SocketProxy string `json:"socket_proxy"`
	HTTPProxy   string `json:"http(s)_proxy"`
	DebugBot    bool   `json:"bot_debug"`
}
