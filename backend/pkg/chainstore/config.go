package chainstore

// Config selects the chain backend for ledger-api.
type Config struct {
	Backend    string `json:"Backend,optional"` // miniledger | fisco
	MiniLedger struct {
		BaseURL string `json:"BaseURL,optional"`
	} `json:"MiniLedger,optional"`
	FISCO FISCOConfig `json:"FISCO,optional"`
}

// FISCOConfig holds FISCO BCOS JSON-RPC and contract addresses.
type FISCOConfig struct {
	JSONRPCURL       string `json:"JSONRPCURL,optional"`
	GroupID          int    `json:"GroupID,optional"`
	ChainID          int64  `json:"ChainID,optional"`
	RegistryContract string `json:"RegistryContract,optional"`
	PrivateKeyPath   string `json:"PrivateKeyPath,optional"`
	ExplorerURL      string `json:"ExplorerURL,optional"`
}
