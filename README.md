# GoBilibiliLogin
 golang实现的bilibili登录

 改自 [login_bili_go](https://github.com/XiaoMiku01/login_bili_go), 适配我的其他程序, 方便复用

## 用法
### 安装
```cmd
go get -u https://github.com/boxie123/GoBilibiliLogin
```

### 示例代码

```go
package main

import (
	"fmt"
	login "github.com/boxie123/GoBilibiliLogin"
)

func main() {
	cookie, csrf, configFilePath := login.Login()

	fmt.Printf("Cookie: %s\ncsrf: %s\ncookie文件存储路径: %s\n", cookie, csrf, configFilePath)
}
```
可在命令行参数中传入来指定文件名，若不传入则默认`bzcookie.json`