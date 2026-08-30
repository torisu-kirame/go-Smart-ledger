package appconfig

import (
	"os"
	"strings"

	nsqmq "github.com/smart-ledger/go-smart-ledger/go-backend/pkg/mq/nsq"
	"github.com/zeromicro/go-zero/rest"
)

// Config is the unified Gin monolith configuration (no service-internal imports).
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
	NSQ       nsqmq.Config `json:"NSQ"`
	BackupDir string       `json:",default=./data/backups"`
	HDWallet  struct {
		Mnemonic string `json:",optional"`
	} `json:"HDWallet"`
	IPFS struct {
		ApiURL  string `json:",optional"`
		Enabled bool   `json:",default=false"`
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

	OpenClaw struct {
		GatewayURL   string `json:",default=http://127.0.0.1:18789"`
		GatewayToken string `json:",optional"`
		AgentModel   string `json:",default=openclaw/default"`
	} `json:"OpenClaw"`
	Agent struct {
		ConfigPath string `json:",default=./data/agent/config"`
	} `json:"Agent"`
	Cors struct {
		AllowedOrigins []string `json:",default=http://localhost:25173"`
	} `json:"Cors"`
}

func (c *Config) ApplyEnv() {
	if v := strings.TrimSpace(os.Getenv("OPENCLAW_GATEWAY_URL")); v != "" {
		c.OpenClaw.GatewayURL = v
	}
	if v := strings.TrimSpace(os.Getenv("OPENCLAW_GATEWAY_TOKEN")); v != "" {
		c.OpenClaw.GatewayToken = v
	}
	if v := strings.TrimSpace(os.Getenv("OPENCLAW_AGENT_MODEL")); v != "" {
		c.OpenClaw.AgentModel = v
	}
	if v := strings.TrimSpace(os.Getenv("AGENT_CONFIG_PATH")); v != "" {
		c.Agent.ConfigPath = v
	}
	if v := strings.TrimSpace(os.Getenv("SL_CHAIN_BACKEND")); v != "" {
		c.Chain.Backend = v
	}
}
