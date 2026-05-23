// Package registry provides etcd-based service registration and discovery (F28).
package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdConfig configures etcd registration.
type EtcdConfig struct {
	Enabled   bool     `json:",optional"`
	Hosts     []string `json:",optional"`
	KeyPrefix string   `json:",default=/smart-ledger/services"`
}

// Endpoint is a registered service instance.
type Endpoint struct {
	Name     string `json:"name"`
	HTTP     string `json:"http,omitempty"`
	GRPC     string `json:"grpc,omitempty"`
	Revision int64  `json:"revision,omitempty"`
}

type record struct {
	HTTP string `json:"http"`
	GRPC string `json:"grpc"`
}

// Register keeps a lease on name -> {http, grpc} until close is called.
func Register(ctx context.Context, cfg EtcdConfig, name, httpAddr, grpcAddr string) (clientv3.LeaseID, func(), error) {
	if !cfg.Enabled || len(cfg.Hosts) == 0 {
		return 0, func() {}, nil
	}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Hosts,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return 0, nil, err
	}
	lease, err := cli.Grant(ctx, 30)
	if err != nil {
		_ = cli.Close()
		return 0, nil, err
	}
	key := cfg.KeyPrefix + "/" + name
	val, _ := json.Marshal(record{HTTP: httpAddr, GRPC: grpcAddr})
	_, err = cli.Put(ctx, key, string(val), clientv3.WithLease(lease.ID))
	if err != nil {
		_ = cli.Close()
		return 0, nil, err
	}
	ch, err := cli.KeepAlive(ctx, lease.ID)
	if err != nil {
		_ = cli.Close()
		return 0, nil, err
	}
	go func() {
		for range ch {
		}
	}()
	closeFn := func() {
		_, _ = cli.Revoke(context.Background(), lease.ID)
		_ = cli.Close()
	}
	return lease.ID, closeFn, nil
}

// Discover lists registered endpoints under prefix.
func Discover(ctx context.Context, cfg EtcdConfig) ([]Endpoint, error) {
	if !cfg.Enabled || len(cfg.Hosts) == 0 {
		return nil, nil
	}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Hosts,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	prefix := cfg.KeyPrefix
	if prefix == "" {
		prefix = "/smart-ledger/services"
	}
	resp, err := cli.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	out := make([]Endpoint, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var rec record
		if err := json.Unmarshal(kv.Value, &rec); err != nil {
			continue
		}
		name := string(kv.Key[len(prefix):])
		if len(name) > 0 && name[0] == '/' {
			name = name[1:]
		}
		out = append(out, Endpoint{
			Name:     name,
			HTTP:     rec.HTTP,
			GRPC:     rec.GRPC,
			Revision: kv.ModRevision,
		})
	}
	return out, nil
}

// ResolveHTTP returns the HTTP base URL for a service name.
func ResolveHTTP(ctx context.Context, cfg EtcdConfig, name string) (string, error) {
	eps, err := Discover(ctx, cfg)
	if err != nil {
		return "", err
	}
	for _, e := range eps {
		if e.Name == name && e.HTTP != "" {
			return e.HTTP, nil
		}
	}
	return "", fmt.Errorf("registry: %s not found", name)
}
