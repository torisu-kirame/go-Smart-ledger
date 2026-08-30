package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/appconfig"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/router/ginadapt"
	authmount "github.com/smart-ledger/go-smart-ledger/go-backend/services/auth/mount"
	gwmount "github.com/smart-ledger/go-smart-ledger/go-backend/services/gateway/mount"
	ledgermount "github.com/smart-ledger/go-smart-ledger/go-backend/services/ledger/mount"
	storagemount "github.com/smart-ledger/go-smart-ledger/go-backend/services/storage/mount"
	"github.com/zeromicro/go-zero/core/conf"
)

var configFile = flag.String("f", "etc/server.yaml", "config file")

func main() {
	flag.Parse()

	var c appconfig.Config
	conf.MustLoad(*configFile, &c)
	c.ApplyEnv()
	if c.Host == "" {
		c.Host = "0.0.0.0"
	}
	if c.Port == 0 {
		c.Port = 28080
	}

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery(), gin.Logger(), corsMiddleware(c.Cors.AllowedOrigins))

	r := ginadapt.New(engine, c.Auth.AccessSecret)

	if _, err := authmount.Register(r, c); err != nil {
		panic(err)
	}
	ledgerCtx, lc, err := ledgermount.Register(r, c)
	if err != nil {
		panic(err)
	}
	defer lc.StopAll(ledgerCtx)

	if _, err := storagemount.Register(r, c); err != nil {
		panic(err)
	}

	gwmount.RegisterAI(r, gwmount.AIOptions{
		OpenClawURL:   c.OpenClaw.GatewayURL,
		OpenClawToken: c.OpenClaw.GatewayToken,
		AgentModel:    c.OpenClaw.AgentModel,
		LedgerURL:     gwmount.SelfURL(c.Port),
		AgentRoot:     c.Agent.ConfigPath,
	})

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	fmt.Printf("Smart Ledger API (Gin monolith) listening on %s\n", addr)
	if err := engine.Run(addr); err != nil {
		panic(err)
	}
}

func corsMiddleware(origins []string) gin.HandlerFunc {
	allow := map[string]bool{}
	for _, o := range origins {
		o = strings.TrimSpace(o)
		if o != "" {
			allow[o] = true
		}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && (len(allow) == 0 || allow[origin] || allow["*"]) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
