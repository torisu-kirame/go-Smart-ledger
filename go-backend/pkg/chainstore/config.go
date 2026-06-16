package chainstore

// Config selects the chain backend for ledger-api.
type Config struct {
	Backend    string `json:"Backend,optional"` // miniledger | fisco
	MiniLedger struct {
		BaseURL string `json:"BaseURL,optional"`
	} `json:"MiniLedger,optional"`
	FISCO FISCOConfig `json:"FISCO,optional"`
}

// FISCOConfig holds FISCO BCOS 3.x JSON-RPC and LedgerRegistry contract settings.
type FISCOConfig struct {
	JSONRPCURL       string `json:"JSONRPCURL,optional"`
	GroupID          string `json:"GroupID,optional"`   // default group0
	ChainID          string `json:"ChainID,optional"`   // default chain0
	NodeName         string `json:"NodeName,optional"`  // empty = any node
	IsSMCrypto       bool   `json:"IsSMCrypto,optional"`
	RegistryContract string `json:"RegistryContract,optional"`
	PrivateKeyHex    string `json:"PrivateKeyHex,optional"`
	PrivateKeyPath   string `json:"PrivateKeyPath,optional"`
	ExplorerURL      string `json:"ExplorerURL,optional"`
}
