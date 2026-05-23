package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	MiniLedger struct {
		BaseURL string `json:",default=http://127.0.0.1:24441"`
	} `json:"MiniLedger"`
	BackupDir string `json:",default=./data/backups"`
	HDWallet  struct {
		Mnemonic string `json:",optional"`
	} `json:"HDWallet"`
	Snowflake struct {
		NodeID int64 `json:",default=2"`
	} `json:"Snowflake"`
	IPFS struct {
		ApiURL  string `json:",optional"`
		Enabled bool   `json:",default=true"`
	} `json:"IPFS"`
}
