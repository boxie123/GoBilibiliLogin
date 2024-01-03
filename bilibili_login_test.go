package gobilibililogin

import (
	"fmt"
	"testing"
)

func TestMyFunction(t *testing.T) {
	cookie_str, csrf, _ := Login()
	fmt.Printf("cookie: %s\ncsrf: %s\n", cookie_str, csrf)

	is_login, _, _ := verifyLogin(cookie_str)

	if !is_login {
		t.Error("登陆失败")
	}
}

func TestGetLoginKeyAndLoginUrl(t *testing.T) {
	loginKey, loginUrl := getLoginKeyAndLoginUrl()
	fmt.Println(loginUrl, loginKey)
}
