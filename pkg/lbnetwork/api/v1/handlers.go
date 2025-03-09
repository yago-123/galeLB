package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yago-123/galelb/pkg/registry"
)

type handler struct {
	registry *registry.NodeRegistry
}

func newHandler(registry *registry.NodeRegistry) *handler {
	return &handler{
		registry: registry,
	}
}

func (h *handler) GetNodesStatus(c *gin.Context) {
	c.Status(http.StatusOK)
}

func (h *handler) GetNode(c *gin.Context) {
	c.Status(http.StatusOK)
}
