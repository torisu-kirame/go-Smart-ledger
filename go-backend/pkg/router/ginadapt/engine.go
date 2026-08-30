package ginadapt

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/authjwt"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/router"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
	xerrors "github.com/zeromicro/x/errors"
)

// Engine mounts standard http.HandlerFunc routes onto Gin.
type Engine struct {
	R         *gin.Engine
	JWTSecret string
}

func New(r *gin.Engine, jwtSecret string) *Engine {
	return &Engine{R: r, JWTSecret: jwtSecret}
}

func (e *Engine) Add(method, path string, h http.HandlerFunc) {
	handlers := []gin.HandlerFunc{}
	if requiresJWT(method, path) && e.JWTSecret != "" {
		handlers = append(handlers, JWT(e.JWTSecret))
	}
	handlers = append(handlers, Wrap(h))
	e.R.Handle(method, path, handlers...)
}

var _ router.Registrar = (*Engine)(nil)

func requiresJWT(method, path string) bool {
	if path == "/api/v1/health" {
		return false
	}
	switch path {
	case "/api/v1/auth/captcha",
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/refresh",
		"/api/v1/auth/logout",
		"/api/v1/auth/health",
		"/api/v1/storage/health":
		return false
	}
	// Public user avatar GET /api/v1/users/:userId/avatar
	if method == http.MethodGet && strings.HasPrefix(path, "/api/v1/users/") && strings.HasSuffix(path, "/avatar") {
		return false
	}
	return true
}

// Wrap injects gin path params into go-zero pathvar context.
func Wrap(h http.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		params := make(map[string]string, len(c.Params))
		for _, p := range c.Params {
			params[p.Key] = p.Value
		}
		req := pathvar.WithVars(c.Request, params)
		h(c.Writer, req)
	}
}

// JWT validates access tokens and sets X-User-Id / X-Username (same as gateway).
func JWT(accessSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := authjwt.AccessFromHeader(c.Request)
		if token == "" {
			httpx.ErrorCtx(c.Request.Context(), c.Writer, xerrors.New(401, "missing access token"))
			c.Abort()
			return
		}
		claims, err := authjwt.Parse(accessSecret, token, authjwt.ClaimTypeAccess)
		if err != nil {
			httpx.ErrorCtx(c.Request.Context(), c.Writer, xerrors.New(401, "invalid or expired access token"))
			c.Abort()
			return
		}
		c.Request.Header.Set("X-User-Id", claims.UserID)
		c.Request.Header.Set("X-Username", claims.Username)
		c.Next()
	}
}
