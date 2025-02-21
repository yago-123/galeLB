package main

import (
	"fmt"
	"hash/crc32"
	"sort"
)

type Hasher func([]byte) uint32

func Crc32Hasher(key []byte) uint32 {
	return crc32.ChecksumIEEE(key)
}

type ConsistentHashing struct {
	ring            []uint32
	nodes           map[uint32]string
	numVirtualNodes int

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
	for i := 0; i < ch.numVirtualNodes; i++ {
		virtualNode := fmt.Sprintf("%s-%d", node, i)
		hash := ch.hasher([]byte(virtualNode))
		ch.ring = append(ch.ring, hash)
		ch.nodes[hash] = node
	}
	sort.Slice(ch.ring, func(i, j int) bool { return ch.ring[i] < ch.ring[j] })
}

func (ch *ConsistentHashing) RemoveNode(node string) {
	for i := 0; i < ch.numVirtualNodes; i++ {
		virtualNode := fmt.Sprintf("%s-%d", node, i)
		hash := ch.hasher([]byte(virtualNode))
		delete(ch.nodes, hash)
	}
	ch.ring = []uint32{}
	for hash := range ch.nodes {
		ch.ring = append(ch.ring, hash)
	}
	sort.Slice(ch.ring, func(i, j int) bool { return ch.ring[i] < ch.ring[j] })
}

func (ch *ConsistentHashing) GetNode(requestKey []byte) string {
	hash := ch.hasher(requestKey)
	i := sort.Search(len(ch.ring), func(i int) bool {
		return ch.ring[i] >= hash
	})
	if i == len(ch.ring) {
		i = 0
	}
	return ch.nodes[ch.ring[i]]
}
