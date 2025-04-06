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
	DNATXDPProgName   = "dnat_prog"
	SNATXDPProgName   = "snat_prog"
)

type xdp struct {
	pubNetInterface  string
	privNetInterface string
	port             int
	logger           *logrus.Logger
}

func newXDP(logger *logrus.Logger, pubNetInterface, privNetInterface string, incomingReqPort int) *xdp {
	return &xdp{
		pubNetInterface:  pubNetInterface,
		privNetInterface: privNetInterface,
		port:             incomingReqPort,
		logger:           logger,
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

	// Retrieve DNAT as SNAT programs from the collection
	progDNAT, found := collection.Programs[DNATXDPProgName]
	if !found {
		return fmt.Errorf("failed to find XDP collection program: %s", DNATXDPProgName)
	}

	progSNAT, found := collection.Programs[SNATXDPProgName]
	if !found {
		return fmt.Errorf("failed to find XDP collection program: %s", SNATXDPProgName)
	}

	// Fetch index of network card based on public interface name
	pubIdxInterface, err := getInterfaceIndex(r.pubNetInterface)
	if err != nil {
		return fmt.Errorf("failed to get index for public network card: %w", err)
	}

	privIdxInterface, err := getInterfaceIndex(r.privNetInterface)
	if err != nil {
		return fmt.Errorf("failed to get index for private network card: %w", err)
	}

	// Attach XDP DNAT program to public network interface
	linkDNAT, err := link.AttachXDP(link.XDPOptions{
		Program:   progDNAT,
		Interface: pubIdxInterface,
	})
	if err != nil {
		return fmt.Errorf("failed to attach XDP link: %w", err)
	}
	defer linkDNAT.Close()

	// Attach XDP SNAT program to private network interface
	linkSNAT, err := link.AttachXDP(link.XDPOptions{
		Program:   progSNAT,
		Interface: privIdxInterface,
	})
	if err != nil {
		return fmt.Errorf("failed to attach XDP link: %w", err)
	}
	defer linkSNAT.Close()

	r.logger.Debugf("XDP program loaded and attached to interface %s (index = %d)", r.pubNetInterface, pubIdxInterface)

	return nil
}

// func (r *xdp) unloadProgram() error {
// 	return nil
// }

func getInterfaceIndex(netInterface string) (int, error) {
	iface, err := net.InterfaceByName(netInterface)
	if err != nil {
		return 0, fmt.Errorf("unable to find interface %s: %w", netInterface, err)
	}
	return iface.Index, nil
}
