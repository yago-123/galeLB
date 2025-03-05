package routing

import (
	"bytes"
	_ "embed"
	"fmt"
	"net"

	"github.com/cilium/ebpf/rlimit"

	"github.com/sirupsen/logrus"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

//go:embed xdp_obj/xdp_router.o
var xdpProg []byte

const (
	RouterXDPProgPath = "pkg/routing/xdp_obj/xdp_router.o"
	RouterXDPProgName = "xdp_router"
)

type xdp struct {
	netInterface string
	port         int
	logger       *logrus.Logger
}

func newXDP(logger *logrus.Logger, netInterface string, incomingReqPort int) *xdp {
	return &xdp{
		netInterface: netInterface,
		port:         incomingReqPort,
		logger:       logger,
	}
}

func (r *xdp) loadProgram() error {
	if err := rlimit.RemoveMemlock(); err != nil {
		return fmt.Errorf("error removing memlock: %w", err)
	}

	// Load spec from the embedded XDP object
	spec, err := ebpf.LoadCollectionSpecFromReader(bytes.NewReader(xdpProg))
	if err != nil {
		return fmt.Errorf("failed to load XDP collection spec: %w", err)
	}

	// Load the XDP object file (ELF)
	collection, err := ebpf.NewCollection(spec)
	if err != nil {
		return fmt.Errorf("failed to create XDP collection: %w", err)
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
		return fmt.Errorf("failed to get interface index: %w", err)
	}

	// Attach XDP program to a network interface
	link, err := link.AttachXDP(link.XDPOptions{
		Program:   prog,
		Interface: idxInterface,
	})
	if err != nil {
		return fmt.Errorf("failed to attach XDP program: %w", err)
	}
	defer link.Close()

	r.logger.Debugf("XDP program loaded and attached to interface %s (index = %d)", r.netInterface, idxInterface)

	return nil
}

func (r *xdp) unloadProgram() error {
	return nil
}

func getInterfaceIndex(netInterface string) (int, error) {
	iface, err := net.InterfaceByName(netInterface)
	if err != nil {
		return 0, fmt.Errorf("unable to find interface %s: %w", netInterface, err)
	}
	return iface.Index, nil
}
