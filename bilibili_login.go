package gobilibililogin

import (
	"fmt"

	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
)

// Login
//
//	@Description: 自动登录 Bilibili 账号并储存 cookie 到文件
//	@return string cookie
//	@return string csrf
//	@return string configFilePath 登录信息文件存储路径
func Login() (string, string, string) {
	configFilePath, isExist := getSettingFilePath()
	var cookie string
	for {
		if isExist {
			_, cookie = readerSettingFile(configFilePath)
		}
		isLogin, data, csrf := verifyLogin(cookie)
		if isLogin {
			uname := data.Get("data.uname").String()
			fmt.Println(uname + "已登录")
			return cookie, csrf, configFilePath
		}
		fmt.Println("未登录,或cookie已过期,请扫码登录")
		loginKey, loginUrl := getLoginKeyAndLoginUrl()
		qrcode := qrcodeTerminal.New()
		qrcode.Get([]byte(loginUrl)).Print()
		successLogin, err := getQRCodeState(loginKey, configFilePath)
		if !successLogin {
			fmt.Println(err)
			continue
		}
		isExist = true
	}
}
