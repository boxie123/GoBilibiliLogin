package gobilibililogin

// ConfigInfo 配置信息
type ConfigInfo struct {
	AccessKey    string `json:"accessKey"`
	Cookie       string `json:"cookie"`
	RefreshToken string `json:"refresh_token"`
}
