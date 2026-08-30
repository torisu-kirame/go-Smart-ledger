package restadapt

import (
	"net/http"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/router"
	"github.com/zeromicro/go-zero/rest"
)

// Server adapts go-zero rest.Server to router.Registrar.
type Server struct {
	S *rest.Server
}

func New(s *rest.Server) router.Registrar {
	return Server{S: s}
}

func (s Server) Add(method, path string, h http.HandlerFunc) {
	s.S.AddRoutes([]rest.Route{{
		Method:  method,
		Path:    path,
		Handler: h,
	}})
}
