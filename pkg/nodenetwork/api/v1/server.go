package v1

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	nodeConfig "github.com/yago-123/galelb/config/node"
	nodeNet "github.com/yago-123/galelb/pkg/nodenetwork"
)

const (
	ServerReadTimeout  = 5 * time.Second
	ServerWriteTimeout = 5 * time.Second
	ServerIdleTimeout  = 10 * time.Second
	MaxHeaderBytes     = 1 << 20

	ServerShutdownTimeout = 5 * time.Second
)

type NodeNetworkAPI struct {
	server *http.Server

	cfg *nodeConfig.Config
}

func New(cfg *nodeConfig.Config, dispatcher *nodeNet.Dispatcher) *NodeNetworkAPI {
	server := &http.Server{
		Addr:           ":5555", // todo(): replace with cfg
		Handler:        setupRouter(dispatcher),
		ReadTimeout:    ServerReadTimeout,
		WriteTimeout:   ServerWriteTimeout,
		IdleTimeout:    ServerIdleTimeout,
		MaxHeaderBytes: MaxHeaderBytes,
	}
	return &NodeNetworkAPI{
		cfg:    cfg,
		server: server,
	}
}

// Start starts the HTTP API server in a BLOCKING manner
func (n *NodeNetworkAPI) Start() error {
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
func (n *NodeNetworkAPI) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), ServerShutdownTimeout)
	defer cancel()

	return n.server.Shutdown(ctx)
}

func setupRouter(dispatcher *nodeNet.Dispatcher) *gin.Engine {
	// todo(): replace with gin.New()
	router := gin.Default()
	handlr := newHandler(dispatcher)

	// GET requests
	router.GET("/status", handlr.GetStatus)

	// POST requests
	router.POST("/start", handlr.PostStart)
	router.POST("/stop", handlr.PostStop)

	return router
}
