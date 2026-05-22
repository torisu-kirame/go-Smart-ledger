package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Upstreams struct {
		Ledger  string `json:",default=http://127.0.0.1:8888"`
		Storage string `json:",default=http://127.0.0.1:8890"`
	} `json:"Upstreams"`
}
