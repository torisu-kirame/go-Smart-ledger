package svc

import (
	"context"
	"fmt"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/grpchsrv"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/registry"
	"github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/internal/config"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

// Lifecycle holds background workers and service registration.
type Lifecycle struct {
	cancel     context.CancelFunc
	etcdClose  func()
	grpcServer *grpc.Server
	EtcdLease  clientv3.LeaseID
}

// StartBackground starts tx retry worker, gRPC health, and etcd registration.
func (ctx *ServiceContext) StartBackground(c config.Config) (*Lifecycle, error) {
	bgCtx, cancel := context.WithCancel(context.Background())
	lc := &Lifecycle{cancel: cancel}

	if c.Discovery.Grpc.Enabled {
		s, err := grpchsrv.Start(c.Host, c.Discovery.Grpc.Port)
		if err != nil {
			cancel()
			return nil, err
		}
		lc.grpcServer = s
	}

	if c.Discovery.Etcd.Enabled {
		httpAddr := c.Discovery.RegisterHTTP
		if httpAddr == "" {
			httpAddr = fmt.Sprintf("http://%s:%d", c.Host, c.Port)
		}
		grpcAddr := c.Discovery.RegisterGRPC
		if grpcAddr == "" && c.Discovery.Grpc.Enabled {
			grpcAddr = fmt.Sprintf("%s:%d", c.Host, c.Discovery.Grpc.Port)
		}
		lease, closeFn, err := registry.Register(bgCtx, c.Discovery.Etcd, "ledger-api", httpAddr, grpcAddr)
		if err != nil {
			lc.StopAll(ctx)
			return nil, err
		}
		lc.EtcdLease = lease
		lc.etcdClose = closeFn
		ctx.EtcdLease = lease
	}

	return lc, nil
}

func (lc *Lifecycle) Stop() {
	if lc.cancel != nil {
		lc.cancel()
	}
}

// StopWithContext stops background workers including NSQ consumers.
func (ctx *ServiceContext) StopQueue() {
	if ctx.Queue != nil {
		ctx.Queue.Stop()
	}
}

func (lc *Lifecycle) StopAll(svcCtx *ServiceContext) {
	lc.Stop()
	if svcCtx != nil {
		svcCtx.StopQueue()
	}
	if lc.etcdClose != nil {
		lc.etcdClose()
	}
	if lc.grpcServer != nil {
		lc.grpcServer.GracefulStop()
	}
}
