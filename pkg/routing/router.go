package routing

import (
	"fmt"

	"github.com/yago-123/galelb/pkg/common"

	lbConfig "github.com/yago-123/galelb/config/lb"
)

type Router struct {
	ring *ring
	xdp  *xdp
}

func New(cfg *lbConfig.Config, numVirtualNodes int) (*Router, error) {
	if numVirtualNodes < 1 {
		return nil, fmt.Errorf("number of virtual nodes cannot be less than 1")
	}

	routerProg := newXDP(cfg.Logger, cfg.Local.NetIfaceClients, cfg.Local.ClientsPort)
	if err := routerProg.loadProgram(); err != nil {
		return nil, fmt.Errorf("failed to load XDP program: %w", err)
	}

	return &Router{
		// todo(): add num virtual nodes to load balancer configuration
		ring: newRing(Crc32Hasher, numVirtualNodes),
		xdp:  routerProg,
	}, nil
}

func (r *Router) AddNode(nodeKey common.AddrKey) {
	r.ring.addNode(nodeKey)
}

func (r *Router) RemoveNode(nodeKey common.AddrKey) {
	r.ring.removeNode(nodeKey)
}
