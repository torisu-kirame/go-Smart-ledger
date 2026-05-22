package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string `json:",default=smart-ledger-access-secret-change-me"`
	} `json:"Auth"`
	Upstreams struct {
		Auth    string `json:",default=http://127.0.0.1:28887"`
		Ledger  string `json:",default=http://127.0.0.1:28888"`
		Storage string `json:",default=http://127.0.0.1:28890"`
	} `json:"Upstreams"`
	Cors struct {
		AllowedOrigins []string `json:",default=http://localhost:25173"`
	} `json:"Cors"`
}
