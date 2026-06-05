package chainstore

// Config selects the chain backend for ledger-api.
type Config struct {
	Backend    string `json:"Backend,optional"` // miniledger | fisco
	MiniLedger struct {
		BaseURL string `json:"BaseURL,optional"`
	} `json:"MiniLedger,optional"`
	FISCO FISCOConfig `json:"FISCO,optional"`
}

// FISCOConfig holds FISCO BCOS 3.x JSON-RPC and contract addresses.
type FISCOConfig struct {
	JSONRPCURL       string `json:"JSONRPCURL,optional"` // default http://127.0.0.1:20200
	GroupID          string `json:"GroupID,optional"`    // default group0
	NodeName         string `json:"NodeName,optional"`   // optional RPC node name
	ChainID          string `json:"ChainID,optional"`    // chain0 (informational)
	RegistryContract string `json:"RegistryContract,optional"`
	PrivateKeyHex    string `json:"PrivateKeyHex,optional"`
	PrivateKeyPath   string `json:"PrivateKeyPath,optional"`
	DisableSsl       bool   `json:"DisableSsl,optional"` // dev chain with disable_ssl=true
	IsSMCrypto       bool   `json:"IsSMCrypto,optional"`
	TLSCaFile        string `json:"TLSCaFile,optional"`
	TLSKeyFile       string `json:"TLSKeyFile,optional"`
	TLSCertFile      string `json:"TLSCertFile,optional"`
	ExplorerURL      string `json:"ExplorerURL,optional"`
}
