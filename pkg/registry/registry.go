package registry

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type node struct {
	continuousHealthChecks uint
	lastHealthCheck        time.Time
}

// NodeRegistry is a struct that keeps track of all nodes that are connected to the load balancer
type NodeRegistry struct {
	// registry is used to keep track of all nodes that are connected to the load balancer. This is used to keep
	// track of nodes to which we should route traffic.
	registry map[string]*node

	// blackList is used to keep track of nodes that have failed health checks.
	blackList map[string]time.Time

	// globalLock is used to prevent race conditions when writing to the node registry. In theory, this lock is
	// not required given that the nodeKey + the nature of gRPC connections makes it impossible for a node struct to
	// generate the same exact connection (same IP + ephemeral port). However, it is a good practice to have a lock for
	// the registry to prevent any unexpected behaviour. In case of this lock being a bottleneck, it could be removed
	// and replaced with a more fine-grained lock (that still, would not be needed at all in theory).
	globalLock sync.RWMutex

	logger *logrus.Logger
}

func New(logger *logrus.Logger) *NodeRegistry {
	return &NodeRegistry{
		registry:  map[string]*node{},
		blackList: map[string]time.Time{},
		logger:    logger,
	}
}

// RegisterNode adds a node to the registry
func (n *NodeRegistry) RegisterNode(nodeKey string) {
	n.globalLock.Lock()
	defer n.globalLock.Unlock()

	if _, ok := n.registry[nodeKey]; !ok {
		n.registry[nodeKey] = &node{}
	}
}

// ReportNewHealthCheck updates the last health check time for a node. Updates info such as todo()
func (n *NodeRegistry) ReportNewHealthCheck(nodeKey string) {
	n.globalLock.Lock()
	defer n.globalLock.Unlock()

	nodeInfo, ok := n.registry[nodeKey]
	if !ok {
		// This case should never happen as we are only calling this function after checking if the node is present
		n.logger.Errorf("unable to load node %s from registry ", nodeKey)
		return
	}

	nodeInfo.lastHealthCheck = time.Now()
	nodeInfo.continuousHealthChecks++
}

func (n *NodeRegistry) ReportNodeFailure(nodeKey string) {
	n.globalLock.Lock()
	defer n.globalLock.Unlock()

	n.logger.Debugf("node %s failed to report health check", nodeKey)

	nodeInfo, ok := n.registry[nodeKey]
	if !ok {
		// This case should never happen as we are only calling this function after checking if the node is present
		n.logger.Errorf("unable to load node %s from registry ", nodeKey)
		return
	}

	nodeInfo.continuousHealthChecks = 0
}
