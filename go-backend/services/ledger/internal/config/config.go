package config

import (
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/chainstore"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/mq/nsq"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/registry"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	// Chain selects authoritative ledger backend (miniledger only).
	Chain struct {
		Backend string `json:",default=miniledger,options=miniledger"`
	} `json:"Chain,optional"`
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
	ExternalAnchor struct {
		Enabled             bool   `json:",default=false"`
		RPCURL              string `json:",optional"`
		ChainID             uint64 `json:",optional"`
		ChainName           string `json:",optional"`
		Contract            string `json:",optional"`
		PrivateKeyHex       string `json:",optional"`
		ExplorerURLTemplate string `json:",optional"`
	} `json:"ExternalAnchor"`
}

// ChainStoreConfig maps service config to pkg/chainstore.
func (c Config) ChainStoreConfig() chainstore.Config {
	cfg := chainstore.Config{
		Backend: c.Chain.Backend,
	}
	cfg.MiniLedger.BaseURL = c.MiniLedger.BaseURL
	return cfg
}
