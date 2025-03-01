package routing

/*
#cgo CFLAGS: -I./pkg/routing
#include "constants.h"
*/
import "C"

import (
	"fmt"
	"sort"
	"sync"

	"github.com/yago-123/galelb/pkg/common"
)

type ring struct {
	ring            []uint32
	nodes           map[uint32]common.AddrKey
	numVirtualNodes int

	lock   sync.RWMutex
	hasher Hasher
}

func newRing(hasher Hasher, numVirtualNodes int) *ring {
	return &ring{
		ring:            make([]uint32, C.MAX_NUMBER_VIRTUAL_NODE_ENTRIES),
		nodes:           make(map[uint32]common.AddrKey),
		numVirtualNodes: numVirtualNodes,
		hasher:          hasher,
	}
}

func (ch *ring) addNode(node common.AddrKey) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	// todo(): check that is not already present?
	// todo(): make sure that does not overflow the ring nor the nodes map

	// Hash the virtual node and persist into the ring
	for i := 0; i < ch.numVirtualNodes; i++ {
		virtualNode := fmt.Sprintf("%s-%d", addrToString(node), i)
		hash := ch.hasher([]byte(virtualNode))

		ch.ring = append(ch.ring, hash)
		ch.nodes[hash] = node
	}

	// Make sure that the ring remains in order
	sort.Slice(ch.ring, func(i, j int) bool { return ch.ring[i] < ch.ring[j] })
}

func (ch *ring) removeNode(node common.AddrKey) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	// Remove the virtual nodes from the map
	for i := 0; i < ch.numVirtualNodes; i++ {
		virtualNode := fmt.Sprintf("%s-%d", addrToString(node), i)
		hash := ch.hasher([]byte(virtualNode))
		delete(ch.nodes, hash)
	}

	// Rebuild the ring
	ch.ring = []uint32{}
	for hash := range ch.nodes {
		ch.ring = append(ch.ring, hash)
	}
	sort.Slice(ch.ring, func(i, j int) bool { return ch.ring[i] < ch.ring[j] })
}

func (ch *ring) getNode(requestKey []byte) string {
	ch.lock.RLock()
	defer ch.lock.RUnlock()

	hash := ch.hasher(requestKey)
	i := sort.Search(len(ch.ring), func(i int) bool {
		return ch.ring[i] >= hash
	})
	if i == len(ch.ring) {
		i = 0
	}
	return addrToString(ch.nodes[ch.ring[i]])
}

func addrToString(addr common.AddrKey) string {
	return fmt.Sprintf("%d:%d", addr.IP, addr.Port)
}
