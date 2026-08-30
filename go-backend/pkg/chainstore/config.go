package chainstore

// Config selects the MiniLedger chain backend for ledger-api.
type Config struct {
	Backend    string `json:"Backend,optional"` // miniledger (only)
	MiniLedger struct {
		BaseURL string `json:"BaseURL,optional"`
	} `json:"MiniLedger,optional"`
}
