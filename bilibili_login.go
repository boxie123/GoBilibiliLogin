package gobilibililogin

import (
	"fmt"

	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	utils "github.com/boxie123/GoBilibiliLogin/bilibili_login_utils"
)

func Login() (string, string) {
	configFilePath, isExsit := utils.GetSettingFilePath()
	var cookie string
	for {
		if isExsit {
			_, cookie = utils.ReaderSettingFile(configFilePath)
		}
		is_login, data, csrf := utils.IsLogin(cookie)
		if is_login {
			uname := data.Get("data.uname").String()
			fmt.Println(uname + "已登录")
			return cookie, csrf
		}
		fmt.Println("未登录,或cookie已过期,请扫码登录")
		fmt.Println("请最大化窗口，以确保二维码完整显示，回车继续")
		fmt.Scanf("%s", "")
		login_key, login_url := utils.GetLoginKeyAndLoginUrl()
		qrcode := qrcodeTerminal.New()
		qrcode.Get([]byte(login_url)).Print()
		fmt.Println("若依然无法扫描，请将以下链接复制到B站打开并确认(任意私信一个人,最好是B站官号，发送链接即可打开)")
		fmt.Println(login_url)
		utils.VerifyLogin(login_key, configFilePath)
		isExsit = true
	}
}