package chainstore

import "fmt"

// New builds the MiniLedger chain Store.
func New(cfg Config) (Store, error) {
	backend := Backend(cfg.Backend)
	if backend == "" {
		backend = BackendMiniLedger
	}
	if backend != BackendMiniLedger {
		return nil, fmt.Errorf("chainstore: unknown backend %q (only miniledger is supported)", cfg.Backend)
	}
	url := cfg.MiniLedger.BaseURL
	if url == "" {
		url = "http://127.0.0.1:24441"
	}
	return newMiniLedger(url), nil
}
