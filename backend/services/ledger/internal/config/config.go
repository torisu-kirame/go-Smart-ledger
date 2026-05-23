package config

import (
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/mq/nsq"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/registry"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	MiniLedger struct {
		BaseURL string `json:",default=http://127.0.0.1:24441"`
	} `json:"MiniLedger"`
	TxQueue struct {
		Enabled     bool   `json:",default=true"`
		PersistPath string `json:",default=./data/txqueue.json"`
		MaxAttempts int    `json:",default=30"`
	} `json:"TxQueue"`
	NSQ nsqmq.Config `json:"NSQ"`
	Discovery struct {
		Etcd registry.EtcdConfig `json:"Etcd"`
		Grpc struct {
			Enabled bool `json:",default=false"`
			Port    int  `json:",default=28898"`
		} `json:"Grpc"`
		RegisterHTTP string `json:",optional"`
		RegisterGRPC string `json:",optional"`
	} `json:"Discovery"`
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
