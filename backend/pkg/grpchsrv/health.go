// Package grpchsrv starts a standard gRPC health server (F28).
package grpchsrv

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// Start listens on host:port and serves gRPC health checks.
func Start(host string, port int) (*grpc.Server, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	s := grpc.NewServer()
	hs := health.NewServer()
	healthpb.RegisterHealthServer(s, hs)
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	hs.SetServingStatus("ledger-api", healthpb.HealthCheckResponse_SERVING)
	go func() {
		_ = s.Serve(ln)
	}()
	return s, nil
}
