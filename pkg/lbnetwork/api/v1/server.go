package v1

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/yago-123/galelb/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/yago-123/galelb/config/lb"
	"github.com/yago-123/galelb/pkg/registry"
)

const (
	ServerReadTimeout  = 5 * time.Second
	ServerWriteTimeout = 5 * time.Second
	ServerIdleTimeout  = 10 * time.Second
	MaxHeaderBytes     = 1 << 20

	ServerShutdownTimeout = 5 * time.Second
)

type LoadBalancerAPI struct {
	server *http.Server

	cfg *lb.Config
}

func New(cfg *lb.Config, registry *registry.NodeRegistry) *LoadBalancerAPI {
	ip, err := util.GetIPv4FromInterface(cfg.PrivateInterface.NetIfacePrivate)
	if err != nil {
		cfg.Logger.Fatalf("failed to get IPv4 address from interface %s: %v", cfg.PrivateInterface.NetIfacePrivate, err)
	}

	server := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", ip, cfg.PrivateInterface.APIPort), // todo(): replace with cfg
		Handler:        setupRouter(registry),
		ReadTimeout:    ServerReadTimeout,
		WriteTimeout:   ServerWriteTimeout,
		IdleTimeout:    ServerIdleTimeout,
		MaxHeaderBytes: MaxHeaderBytes,
	}

	return &LoadBalancerAPI{
		cfg:    cfg,
		server: server,
	}
}

// Start starts the HTTP API server in a BLOCKING manner
func (n *LoadBalancerAPI) Start() error {
	err := n.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		n.cfg.Logger.Infof("HTTP API server stopped successfully")
		return nil
	}

	if err != nil {
		return err
	}

	return nil
}

// Stop stops the HTTP API server
func (n *LoadBalancerAPI) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), ServerShutdownTimeout)
	defer cancel()

	return n.server.Shutdown(ctx)
}

func setupRouter(registry *registry.NodeRegistry) *gin.Engine {
	router := gin.Default() // todo(): replace with gin.New()
	handlr := newHandler(registry)

	router.GET("/nodes", handlr.GetNodesStatus)
	router.GET("/nodes/:id", handlr.GetNode)

	return router
}
