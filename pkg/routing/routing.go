package routing

import (
	_ "embed"
	"fmt"
	"github.com/cilium/ebpf/rlimit"
	"net"

	"github.com/sirupsen/logrus"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

const (
	RouterXDPProgPath = "pkg/routing/xdp_obj/xdp_router.o"
	RouterXDPProgName = "xdp_router"
)

type router struct {
	netInterface string
	logger       *logrus.Logger
}

func New(logger *logrus.Logger, netInterface string) *router {
	return &router{
		netInterface: netInterface,
		logger:       logger,
	}
}

func (r *router) LoadRouter() error {
	if err := rlimit.RemoveMemlock(); err != nil {
		return fmt.Errorf("error removing memlock: %v", err)
	}

	// Load the XDP object file (ELF)
	// todo(): this should be loaded with embed instead
	collection, err := ebpf.LoadCollection(RouterXDPProgPath)
	if err != nil {
		return fmt.Errorf("failed to load XDP collection: %v", err)
	}
	defer collection.Close()

	// Retrieve program loaded
	prog, found := collection.Programs[RouterXDPProgName]
	if !found {
		return fmt.Errorf("failed to find XDP collection program: %s", RouterXDPProgName)
	}

	// Fetch index from string interface
	idxInterface, err := getInterfaceIndex(r.netInterface)
	if err != nil {
		return fmt.Errorf("failed to get interface index: %v", err)
	}

	// Attach XDP program to a network interface
	link, err := link.AttachXDP(link.XDPOptions{
		Program:   prog,
		Interface: idxInterface,
	})
	if err != nil {
		return fmt.Errorf("failed to attach XDP program: %v", err)
	}
	defer link.Close()

	r.logger.Debugf("XDP program loaded and attached to interface %s (index = %d)", r.netInterface, idxInterface)

	return nil
}

func (r *router) UnloadRouter() error {
	return nil
}

func (r *router) UpdateRing() {

}

func getInterfaceIndex(netInterface string) (int, error) {
	iface, err := net.InterfaceByName(netInterface)
	if err != nil {
		return 0, fmt.Errorf("unable to find interface %s: %w", netInterface, err)
	}
	return iface.Index, nil
}
