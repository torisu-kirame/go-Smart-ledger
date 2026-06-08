package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/chainstore"
)

func main() {
	miniledgerURL := flag.String("miniledger", "http://127.0.0.1:24441", "MiniLedger base URL")
	fiscoRPC := flag.String("fisco-rpc", "http://127.0.0.1:20200", "FISCO BCOS 3.0 JSON-RPC URL")
	registry := flag.String("registry", "", "LedgerRegistry contract address (required unless -dry-run with -verify=false)")
	groupID := flag.String("group", "group0", "FISCO group ID")
	chainID := flag.String("chain-id", "chain0", "FISCO chain ID")
	privateKey := flag.String("private-key", "", "hex private key for FISCO writes")
	privateKeyFile := flag.String("private-key-file", "", "file containing FISCO private key hex")
	disableSsl := flag.Bool("disable-ssl", false, "deprecated: set disable_ssl on FISCO node config.ini instead")
	dryRun := flag.Bool("dry-run", false, "scan source only, do not write")
	verify := flag.Bool("verify", false, "after write, compare each key on FISCO")
	limit := flag.Int("limit", 0, "max keys to migrate (0 = all)")
	verbose := flag.Bool("verbose", false, "log every key")
	flag.Parse()

	if *registry == "" && !*dryRun {
		log.Fatal("-registry is required for live migration")
	}
	if *privateKey == "" && *privateKeyFile == "" && !*dryRun {
		log.Fatal("-private-key or -private-key-file required for live migration")
	}
	if *privateKeyFile != "" {
		raw, err := os.ReadFile(*privateKeyFile)
		if err != nil {
			log.Fatalf("read private key file: %v", err)
		}
		*privateKey = string(raw)
	}

	src, err := chainstore.New(chainstore.Config{
		Backend: string(chainstore.BackendMiniLedger),
		MiniLedger: struct {
			BaseURL string `json:"BaseURL,optional"`
		}{BaseURL: *miniledgerURL},
	})
	if err != nil {
		log.Fatalf("miniledger store: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)
	defer cancel()

	var dst chainstore.Store
	if !*dryRun {
		if *disableSsl {
			log.Println("note: -disable-ssl is ignored; use node config.ini disable_ssl=true")
		}
		dst, err = chainstore.New(chainstore.Config{
			Backend: string(chainstore.BackendFISCO),
			FISCO: chainstore.FISCOConfig{
				JSONRPCURL:       *fiscoRPC,
				GroupID:          *groupID,
				ChainID:          *chainID,
				RegistryContract: *registry,
				PrivateKeyHex:    *privateKey,
			},
		})
		if err != nil {
			log.Fatalf("fisco store: %v", err)
		}
		if err := dst.Ping(ctx); err != nil {
			log.Fatalf("fisco ping: %v", err)
		}
	}

	if err := src.Ping(ctx); err != nil {
		log.Fatalf("miniledger ping: %v", err)
	}

	opt := chainstore.MigrateOptions{
		DryRun: *dryRun,
		Verify: *verify,
		Limit:  *limit,
		OnProgress: func(done, total int, key string) {
			if *verbose || done == total || done%50 == 0 {
				log.Printf("[%d/%d] %s", done, total, key)
			}
		},
	}

	res, err := chainstore.MigrateMiniLedgerToFISCO(ctx, src, dst, opt)
	raw, _ := json.MarshalIndent(res, "", "  ")
	fmt.Println(string(raw))
	if err != nil {
		log.Fatal(err)
	}
	if *dryRun {
		log.Printf("dry-run complete: would migrate %d keys", res.Total)
	} else {
		log.Printf("migration complete: written=%d failed=%d verified=%d", res.Written, res.Failed, res.Verified)
	}
}
