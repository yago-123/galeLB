package nodenetwork

import (
	"context"
	"fmt"
	"sync"
	"time"

	nodeConfig "github.com/yago-123/galelb/config/node"
	v1Consensus "github.com/yago-123/galelb/pkg/consensus/v1"
)

const (
	GetConfigTimeout           = 5 * time.Second
	HealthCheckIntervalDivisor = 2
)

type Status string

const (
	StatusRunning Status = "running"
	StatusStopped        = "stopped"
)

type Target struct {
	IP   string
	Port int
}

func (t *Target) String() string {
	return fmt.Sprintf("%s:%d", t.IP, t.Port)
}

// Dispatcher contains the logic that determines how and when to dispatch messages to the load balancers. Dispatcher
// initializes N connections to N load balancers to todo()
type Dispatcher struct {
	targets map[string]Target
	status  Status
	lock    sync.RWMutex

	generalCtx    context.Context
	generalCancel context.CancelFunc

	cfg *nodeConfig.Config
}

func NewDispatcher(cfg *nodeConfig.Config, targets map[string]Target) *Dispatcher {
	return &Dispatcher{
		targets: targets,
		status:  StatusStopped,
		cfg:     cfg,
	}
}

func (d *Dispatcher) Start() error {
	var wg sync.WaitGroup

	d.lock.Lock()
	// If the dispatcher is already running, return
	if d.status == StatusRunning {
		d.lock.Unlock()
		return fmt.Errorf("dispatcher is already running")
	}

	d.status = StatusRunning
	d.lock.Unlock()

	// Update the status of the dispatcher once the function returns
	defer func() {
		d.lock.Lock()
		d.status = StatusStopped
		d.lock.Unlock()
	}()

	// Initialize context and cancel function for the dispatcher goroutines, so that they can be stopped from the
	// Stop method anytime and update the status accordingly
	d.generalCtx, d.generalCancel = context.WithCancel(context.Background())

	// Start a new goroutine for each target
	if err := d.startDispatchers(&wg); err != nil {
		return err
	}

	wg.Wait()

	return nil
}

func (d *Dispatcher) Stop() error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.status == StatusStopped {
		return fmt.Errorf("dispatcher is already stopped")
	}

	d.generalCancel()
	d.status = StatusStopped

	return nil
}

func (d *Dispatcher) Status() Status {
	d.lock.RLock()
	defer d.lock.RUnlock()

	return d.status
}

// startDispatchers starts a goroutine for each target in the dispatcher
func (d *Dispatcher) startDispatchers(wg *sync.WaitGroup) error {
	for k, target := range d.targets {
		d.cfg.Logger.Infof("starting dispatcher for %s", k)

		client, err := NewClient(d.cfg.Logger, target.IP, target.Port)
		if err != nil {
			// todo(): if we don't want to keep tracking of failed report health loops, we should return this with an
			// todo(): error so that we can ensure that once startDispatchers returns, all health loops are running "forever"
			return fmt.Errorf("failed to create client for target %s:%d: %w", target.IP, target.Port, err)
		}

		executionCfg, err := d.fetchConfig(client)
		if err != nil {
			// todo(): if we don't want to keep tracking of failed report health loops, we should return this with an
			// todo(): error so that we can ensure that once startDispatchers returns, all health loops are running "forever"
			return fmt.Errorf("failed to fetch config for target %s:%d: %w", target.IP, target.Port, err)
		}

		normalizedTime := time.Duration(executionCfg.GetHealthCheckTimeout()) * time.Nanosecond
		healthPeriod := normalizedTime / HealthCheckIntervalDivisor

		wg.Add(1)
		go d.reportHealthLoop(wg, client, target, normalizedTime, healthPeriod)
	}

	return nil
}

// fetchConfig fetches the configuration from the load balancer
func (d *Dispatcher) fetchConfig(client *Client) (*v1Consensus.ConfigResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), GetConfigTimeout)
	defer cancel()

	executionCfg, err := client.GetConfig(ctx)
	if err != nil {
		d.cfg.Logger.Errorf("failed to get config: %v", err)
		return nil, err
	}

	return executionCfg, nil
}

// reportHealthLoop is a goroutine that reports the health status of the node to the load balancer target
func (d *Dispatcher) reportHealthLoop(wg *sync.WaitGroup, client *Client, t Target, timeout, period time.Duration) {
	defer wg.Done()

	for {
		select {
		case <-d.generalCtx.Done():
			// If the dispatcher is stopped, return
			return
		default:
			// Otherwise, report health status
			ctxTimeout, cancel := context.WithTimeout(context.Background(), timeout)
			if err := client.ReportHealthStatus(ctxTimeout, &v1Consensus.HealthStatus{
				Service: "gale-node",
				Status:  uint32(v1Consensus.Serving),
				Message: "Serving requests goes brrrrr",
			}); err != nil {
				d.cfg.Logger.Errorf("failed to report health status: %v", err)
			}
			cancel()

			d.cfg.Logger.Debugf("reported health status to %s:%d", t.IP, t.Port)
			<-time.NewTimer(period).C
		}
	}
}
