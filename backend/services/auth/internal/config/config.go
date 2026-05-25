package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret  string `json:",default=smart-ledger-access-secret-change-me"`
		RefreshSecret string `json:",default=smart-ledger-refresh-secret-change-me"`
		AccessExpire  int64  `json:",default=900"`
		RefreshExpire int64  `json:",default=604800"`
		CookieSecure  bool   `json:",default=false"`
		CookieDomain  string `json:",optional"`
	} `json:"Auth"`
	// Database 外部 MySQL，见 etc/auth-api.yaml（不在 Compose 内托管）
	Database struct {
		DataSource string
	} `json:"Database"`
	Avatar struct {
		Dir string `json:",default=data/avatars"`
	} `json:"Avatar"`
	TeamChat struct {
		Dir string `json:",default=data/teamchat"`
	} `json:"TeamChat"`
	Snowflake struct {
		NodeID int64 `json:",default=1"`
	} `json:"Snowflake"`
	Users []struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"Users,optional"`
}
