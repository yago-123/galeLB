package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	nodeNet "github.com/yago-123/galelb/pkg/nodenetwork"
)

type handler struct {
	dispatcher *nodeNet.Dispatcher
}

func newHandler(dispatcher *nodeNet.Dispatcher) *handler {
	return &handler{
		dispatcher: dispatcher,
	}
}

// @Summary Get node status
// @Description Retrieve status of the node by checking the status of the dispatcher
// @ID get-status
// @Produce  json
// @Success 200 {object} StatusResponse
// @Router /status [get]
func (h *handler) GetStatus(c *gin.Context) {
	status := h.dispatcher.Status()

	c.JSON(http.StatusOK, StatusResponse{
		Status: string(status),
	})
}

// @Summary Start node network dispatcher
// @Description Send order to dispatcher for starting the node network requests
// @ID post-start
// @Success 200
// @Router /start [post]
func (h *handler) PostStart(c *gin.Context) {
	_ = h.dispatcher.Start()

	c.Status(http.StatusOK)
}

// @Summary Stop node network dispatcher
// @Description Send order to dispatcher for stopping the node network requests
// @ID post-stop
// @Success 200
// @Router /stop [post]
func (h *handler) PostStop(c *gin.Context) {
	_ = h.dispatcher.Stop()

	c.Status(http.StatusOK)
}
