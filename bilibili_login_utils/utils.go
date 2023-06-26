package bilibili_login_utils

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

func GetLoginKeyAndLoginUrl() (string, string) {
	url := "https://passport.bilibili.com/qrcode/getLoginUrl"
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
	login_key := data.Get("data.oauthKey").String()
	login_url := data.Get("data.url").String()
	return login_key, login_url
}

func getLiveBuvid() string {
	url := "https://api.live.bilibili.com/gift/v3/live/gift_config"
	client := http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	SetCookie := resp.Header.Get("Set-Cookie")
	reg := re.MustCompile(`LIVE_BUVID=(AUTO[0-9]+)`)
	live_buvid := reg.FindStringSubmatch(SetCookie)[1]
	return live_buvid
}

func VerifyLogin(login_key string, filePath string) {
	for {
		apiUrl := "https://passport.bilibili.com/qrcode/getLoginInfo"
		client := http.Client{}
		params := url.Values{"oauthKey": {login_key}}
		req, _ := http.NewRequest("POST", apiUrl, strings.NewReader(params.Encode()))
		req.Header.Set("User-Agent", UserAgent)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		data := gjson.ParseBytes(body)
		if data.Get("status").Bool() {
			url := data.Get("data.url").String()
			reg := re.MustCompile(`DedeUserID=(\d+)&DedeUserID__ckMd5=([0-9a-zA-Z]+)&Expires=(\d+)&SESSDATA=([0-9a-zA-Z%]+)&bili_jct=([0-9a-zA-Z]+)&`)
			cookie := make(map[string]string)
			cookie["DedeUserID"] = reg.FindStringSubmatch(url)[1]
			cookie["DedeUserID__ckMd5"] = reg.FindStringSubmatch(url)[2]
			cookie["Expires"] = reg.FindStringSubmatch(url)[3]
			cookie["SESSDATA"] = reg.FindStringSubmatch(url)[4]
			cookie["bili_jct"] = reg.FindStringSubmatch(url)[5]
			cookie["LIVE_BUVID"] = getLiveBuvid()
			cookie_content := "DedeUserID=" + cookie["DedeUserID"] + "; DedeUserID__ckMd5=" + cookie["DedeUserID__ckMd5"] + "; Expires=" + cookie["Expires"] + "; SESSDATA=" + cookie["SESSDATA"] + "; bili_jct=" + cookie["bili_jct"] + "; LIVE_BUVID=" + cookie["LIVE_BUVID"]
			configInfo := ConfigInfo{Cookie: cookie_content}
			jsonData, err := json.MarshalIndent(configInfo, "", "    ")
			if err != nil {
				fmt.Println("Error marshalling JSON:", err)
				return
			}

			err = os.WriteFile(filePath, jsonData, 0644)
			if err != nil {
				panic(err)
			}
			s := fmt.Sprintf("扫码成功, cookie如下,已自动保存在当前目录下 %v 文件:", filePath)
			fmt.Println(s)
			fmt.Println(string(cookie_content))
			break
		}
		time.Sleep(time.Second * 3)
	}
}

func IsLogin(cookie string) (bool, gjson.Result, string) {
	url := "https://api.bilibili.com/x/web-interface/nav"

	client := http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Cookie", cookie)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	data := gjson.ParseBytes(body)

	isLogin := data.Get("code").Int() == 0
	var csrf string
	if isLogin {
		reg := re.MustCompile(`bili_jct=([0-9a-zA-Z]+);`)
		csrf = reg.FindStringSubmatch(cookie)[1]
	}
	return data.Get("code").Int() == 0, data, csrf
}

// 读取配置文件
func ReaderSettingFile(filePath string) (string, string) {
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
func GetSettingFilePath() (string, bool) {
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
