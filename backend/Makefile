.PHONY: tidy build test run-ledger run-storage run-gateway run-miniledger

tidy:
	go mod tidy

build:
	go build -o bin/ledger-api.exe ./services/ledger
	go build -o bin/storage-api.exe ./services/storage
	go build -o bin/gateway-api.exe ./services/gateway

test:
	go test ./...

run-miniledger:
	cd infra/miniledger && npm install && npm run start

run-ledger:
	cd services/ledger && go run ledger.go -f etc/ledger-api.yaml

run-storage:
	cd services/storage && go run storage.go -f etc/storage-api.yaml

run-gateway:
	cd services/gateway && go run gateway.go -f etc/gateway-api.yaml
