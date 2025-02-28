package routing

import (
	"fmt"

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

	xdpProg := newXDP(cfg.Logger, cfg.Local.NetIfaceClients, cfg.Local.ClientsPort)
	if err := xdpProg.loadProgram(); err != nil {
		return nil, fmt.Errorf("failed to load XDP program: %w", err)
	}

	return &Router{
		// todo(): add num virtual nodes to load balancer configuration
		ring: newRing(Crc32Hasher, 5),
		xdp:  xdpProg,
	}, nil
}

func (r *Router) AddNode(nodeKey string, ip string, port int) {
	r.ring.addNode(nodeKey)
}

func (r *Router) RemoveNode(nodeKey string) {
	r.ring.removeNode(nodeKey)
}
