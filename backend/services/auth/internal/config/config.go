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
	MySQL struct {
		DataSource string `json:",optional"`
	} `json:"MySQL"`
	Users []struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"Users,optional"`
}
