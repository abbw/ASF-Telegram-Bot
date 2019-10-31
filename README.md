# ASF-Telegram-Bot
**Version 0.9.4**  
通过ASF的IPC接口接入的Telegram机器人  
实现「**通过Telegram控制自己的ASF**」  
Updated at 2019-10-31  
- [x] 无配置文件按默认值自动生成
- [x] 检查IPC配置正确性
- [x] 检查tg bot配置正确性
- [x] 非主人指令返回chat id
- [x] 主人指令执行ASF
- [x] 优化控制台
- [x] 处理Panic异常
- [x] 解决英文环境的控制台部分内容显示乱码
- [x] 0.9.4 加入http/https/socket5代理

## Installation 下载链接
### Windows Version
直接运行ASFbot.exe「[Windows版](https://github.com/rakuyo42/ASF-Telegram-Bot/releases/download/v0.9.4/ASF_Tg_Bot.exe)」  
### Linux Version
`chmod 775 ASFbot`给予运行权限后`./ASFbot`执行「[Linux版](https://github.com/rakuyo42/ASF-Telegram-Bot/releases/download/v0.9.4/asf_tg_bot)」  
（也可以用win10的Powershell运行）
### QQ机器人
[酷Q插件.cpk](https://github.com/rakuyo42/ASF-Telegram-Bot/releases/download/v0.9.4/ink.ews.steamhelper.cpk)
### 安全警告
如果你不放心可以下载源码自己编译，推荐有条件的都自己编译  
保持思考，拒绝盲从。  
酷q插件的源码在[另一个仓库](https://github.com/rakuyo42/CoolQ-cpks/blob/master/dev/ink.ews.steamhelper/app.go)

## Quick start
首次启动，生成配置文件后按照控制台提示修改配置  
或者  
先退出，直接编辑同目录下的config.json配置文件再重新启动 

## config.json配置项
```json
{
    "tg_bot_token": "你的机器人token",
    "tg_chat_id": 你的Telegram账号的chatid(注意此项没有用引号括起来，此项应该是一串不带引号的纯数字),
    "asf_ipc_url": "你(已经开启了IPC)的ASF所在的IP",
    "asf_ipc_password": "你设置的IPC密码，如果你设置了的话",
    "socket_proxy": "如果你的ASF无法连接到telegram，如果是本地代理可以只填端口号",
    "http(s)_proxy": "同上，但如果是https注意要用完整url[https://你的代理地址:你的端口]",
    "bot_debug": true or false(是否打开机器人的debug信息，打开了控制台会很乱，默认关闭)
}
```
