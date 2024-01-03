package gobilibililogin

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	re "regexp"
	"strings"
	"time"

	gjson "github.com/tidwall/gjson"
)

const UserAgent = `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.99 Safari/537.36 Edg/97.0.1072.69`

// getLoginKeyAndLoginUrl
//
//	@Description: 获取二维码内容和密钥
//	@return string 密钥
//	@return string 二维码链接
func getLoginKeyAndLoginUrl() (string, string) {
	url := "https://passport.bilibili.com/x/passport-login/web/qrcode/generate"
	client := http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	data := gjson.ParseBytes(body)
	loginKey := data.Get("data.qrcode_key").String()
	loginUrl := data.Get("data.url").String()
	return loginKey, loginUrl
}

// 获取二维码状态
func getQRCodeState(loginKey string, filePath string) (bool, error) {
	for {
		apiUrl := "https://passport.bilibili.com/x/passport-login/web/qrcode/poll"
		client := http.Client{}
		req, _ := http.NewRequest("GET", apiUrl+fmt.Sprintf("?qrcode_key=%s", loginKey), nil)
		req.Header.Set("User-Agent", UserAgent)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := client.Do(req)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		data := gjson.ParseBytes(body)
		switch data.Get("data.code").Int() {
		case 0:
			cookieUrl := data.Get("data.url").String()
			parsedUrl, err := url.Parse(cookieUrl)
			if err != nil {
				return false, err
			}
			cookieContentList := strings.Split(parsedUrl.RawQuery, "&")
			cookieContent := ""
			for _, cookie := range cookieContentList[:len(cookieContentList)-1] {
				cookieContent = cookieContent + cookie + ";"
			}
			cookieContent = strings.TrimSuffix(cookieContent, ";")
			configInfo := ConfigInfo{
				Cookie:       cookieContent,
				RefreshToken: data.Get("data.refresh_token").String(),
			}
			jsonData, err := json.MarshalIndent(configInfo, "", "    ")
			if err != nil {
				return false, err
			}
			err = os.WriteFile(filePath, jsonData, 0644)
			if err != nil {
				return false, err
			}
			s := fmt.Sprintf("扫码成功, 已自动保存在当前目录下 %v 文件:", filePath)
			fmt.Println(s)
			return true, nil
		case 86038:
			fmt.Println("二维码已失效，正在重新生成")
			return false, fmt.Errorf("二维码失效")
		case 86090:
			fmt.Println("已扫码，请确认")
		case 86101:
		default:
			return false, fmt.Errorf("未知code: %d", data.Get("data.code").Int())
		}
		time.Sleep(time.Second * 3)
	}
}

// 验证 cookie 可用性
func verifyLogin(cookie string) (bool, gjson.Result, string) {
	u := "https://api.bilibili.com/x/web-interface/nav"

	client := http.Client{}
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Cookie", cookie)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	data := gjson.ParseBytes(body)

	isLogin := data.Get("data.isLogin").Bool()
	var csrf string
	if isLogin {
		reg := re.MustCompile(`bili_jct=([0-9a-zA-Z]+)`)
		csrf = reg.FindStringSubmatch(cookie)[1]
	}
	return isLogin, data, csrf
}

// 读取配置文件
func readerSettingFile(filePath string) (string, string) {
	var ConfigData, _ = os.ReadFile(filePath)
	var configContent = ConfigInfo{}

	err := json.Unmarshal(ConfigData, &configContent)
	if err != nil {
		panic("读取登录信息失败")
	}

	var cookie = configContent.Cookie
	var accessKey = configContent.AccessKey

	return accessKey, cookie
}

// 获取配置文件路径, 并判断路径是否存在
func getSettingFilePath() (string, bool) {
	var FilePath string
	if len(os.Args) <= 1 {
		log.Println("未选择配置文件, 默认为bzcookie.json")
		FilePath = filepath.Join(".", "bzcookie.json")
	} else {
		FilePath = os.Args[len(os.Args)-1]
	}
	_, err := os.Lstat(FilePath)
	if err != nil {
		log.Printf("[%v]不存在\n", FilePath)
		FilePath = filepath.Join(".", "bzcookie.json")
		_, err = os.Lstat(FilePath)
		if err != nil {
			log.Println("[bzcookie.json]也不存在")
			return FilePath, false
		}
		log.Println("[bzcookie.json]存在, 将使用其中登录信息")
		return FilePath, true
	}
	log.Printf("配置文件:[%v]\n", FilePath)
	return FilePath, true
}
