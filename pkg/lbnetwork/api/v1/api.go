package v1

import (
	"github.com/yago-123/galelb/config/lb"
	"github.com/yago-123/galelb/pkg/registry"
)

type LoadBalancerAPI struct {
	registry *registry.NodeRegistry

	cfg *lb.Config
}

func New(cfg *lb.Config, registry *registry.NodeRegistry) *LoadBalancerAPI {
	return &LoadBalancerAPI{
		cfg:      cfg,
		registry: registry,
	}
}

func (l *LoadBalancerAPI) Start() {

}

func (l *LoadBalancerAPI) Stop() {

}

// add router with options for retrieving the current registry of nodes
