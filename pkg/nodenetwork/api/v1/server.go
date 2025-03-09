package v1

import (
	"github.com/gin-gonic/gin"
	nodeConfig "github.com/yago-123/galelb/config/node"
	nodeNet "github.com/yago-123/galelb/pkg/nodenetwork"
	"net/http"
	"time"
)

const (
	ServerReadTimeout  = 5 * time.Second
	ServerWriteTimeout = 5 * time.Second
	ServerIdleTimeout  = 10 * time.Second
	MaxHeaderBytes     = 1 << 20
)

type NodeNetworkAPI struct {
	dispatcher *nodeNet.Dispatcher

	server *http.Server

	cfg *nodeConfig.Config
}

func New(cfg *nodeConfig.Config, dispatcher *nodeNet.Dispatcher) *NodeNetworkAPI {
	server := &http.Server{
		Addr:           ":5555", // todo(): replace with cfg
		Handler:        setupRouter(),
		ReadTimeout:    ServerReadTimeout,
		WriteTimeout:   ServerWriteTimeout,
		IdleTimeout:    ServerIdleTimeout,
		MaxHeaderBytes: MaxHeaderBytes,
	}
	return &NodeNetworkAPI{
		cfg:        cfg,
		dispatcher: dispatcher,
		server:     server,
	}
}

func (n *NodeNetworkAPI) Start() {

}

func (n *NodeNetworkAPI) Stop() {

}

// add router with options for starting / stopping dispatcher

func setupRouter() *gin.Engine {
	// todo(): replace with gin.New()
	router := gin.Default()

	router.GET("/status", func(c *gin.Context) {

	})

	router.POST("/start", func(c *gin.Context) {

	})

	router.POST("/stop", func(c *gin.Context) {

	})

	return router
}
