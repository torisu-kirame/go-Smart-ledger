package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	MiniLedger struct {
		BaseURL string `json:",default=http://127.0.0.1:4441"`
	} `json:"MiniLedger"`
}
