package chainstore

import "fmt"

// New builds the configured chain Store. Empty Backend defaults to miniledger.
func New(cfg Config) (Store, error) {
	backend := Backend(cfg.Backend)
	if backend == "" {
		backend = BackendMiniLedger
	}
	switch backend {
	case BackendMiniLedger:
		url := cfg.MiniLedger.BaseURL
		if url == "" {
			url = "http://127.0.0.1:24441"
		}
		return newMiniLedger(url), nil
	case BackendFISCO:
		return NewFISCO(cfg.FISCO)
	default:
		return nil, fmt.Errorf("chainstore: unknown backend %q", cfg.Backend)
	}
}
