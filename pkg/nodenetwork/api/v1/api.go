package v1

import (
	nodeConfig "github.com/yago-123/galelb/config/node"
	nodeNet "github.com/yago-123/galelb/pkg/nodenetwork"
)

type NodeNetworkAPI struct {
	dispatcher *nodeNet.Dispatcher

	cfg *nodeConfig.Config
}

func New(cfg *nodeConfig.Config, dispatcher *nodeNet.Dispatcher) *NodeNetworkAPI {
	return &NodeNetworkAPI{
		cfg:        cfg,
		dispatcher: dispatcher,
	}
}

func (n *NodeNetworkAPI) Start() {

}

func (n *NodeNetworkAPI) Stop() {

}

// add router with options for starting / stopping dispatcher
