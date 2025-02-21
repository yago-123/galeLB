package ring

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

type ConsistentHashing struct {
	ring            []uint32
	nodes           map[uint32]string
	numVirtualNodes int

	lock   sync.RWMutex
	hasher Hasher
}

func New(hasher Hasher, numVirtualNodes int) (*ConsistentHashing, error) {
	if numVirtualNodes < 1 {
		return nil, fmt.Errorf("number of virtual nodes cannot be less than 1")
	}

	return &ConsistentHashing{
		ring:            []uint32{},
		nodes:           make(map[uint32]string),
		numVirtualNodes: numVirtualNodes,
		hasher:          hasher,
	}, nil
}

func (ch *ConsistentHashing) AddNode(node string) {
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

func (ch *ConsistentHashing) RemoveNode(node string) {
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

func (ch *ConsistentHashing) GetNode(requestKey []byte) string {
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
