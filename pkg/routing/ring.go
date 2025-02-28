package routing

import (
	"fmt"
	"hash/crc32"
	"sort"
	"sync"
)

type Hasher func([]byte) uint32

func Crc32Hasher(key []byte) uint32 {
	return crc32.ChecksumIEEE(key)
}

type ring struct {
	ring            []uint32
	nodes           map[uint32]string
	numVirtualNodes int

	lock   sync.RWMutex
	hasher Hasher
}

func newRing(hasher Hasher, numVirtualNodes int) *ring {
	return &ring{
		ring:            []uint32{},
		nodes:           make(map[uint32]string),
		numVirtualNodes: numVirtualNodes,
		hasher:          hasher,
	}
}

func (ch *ring) addNode(node string) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	// Hash the virtual node and persist into the ring
	for i := 0; i < ch.numVirtualNodes; i++ {
		virtualNode := fmt.Sprintf("%s-%d", node, i)
		hash := ch.hasher([]byte(virtualNode))

		ch.ring = append(ch.ring, hash)
		ch.nodes[hash] = node
	}

	// Make sure that the ring remains in order
	sort.Slice(ch.ring, func(i, j int) bool { return ch.ring[i] < ch.ring[j] })
}

func (ch *ring) removeNode(node string) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	// Remove the virtual nodes from the map
	for i := 0; i < ch.numVirtualNodes; i++ {
		virtualNode := fmt.Sprintf("%s-%d", node, i)
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
	return ch.nodes[ch.ring[i]]
}
