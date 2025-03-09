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

	cfg *nodeConfig.Config
}

func NewDispatcher(cfg *nodeConfig.Config, targets map[string]Target) *Dispatcher {
	return &Dispatcher{
		targets: targets,
		cfg:     cfg,
	}
}

// todo(): adjust behaviour to fit dispatcher characteristics
func (d *Dispatcher) Start() {
	var wg sync.WaitGroup

	for k, v := range d.targets {
		d.cfg.Logger.Infof("starting dispatcher for %s", k)

		client, err := NewClient(d.cfg.Logger, v.IP, v.Port)
		if err != nil {
			d.cfg.Logger.Errorf("failed to create client: %v", err)
			continue
		}

		ctxConfig, cancelConfig := context.WithTimeout(context.Background(), GetConfigTimeout)
		defer cancelConfig()

		executionCfg, err := client.GetConfig(ctxConfig)
		if err != nil {
			d.cfg.Logger.Errorf("failed to get config: %v", err)
			continue
		}

		// todo(): define better times and intervals
		normalizedTime := time.Duration(executionCfg.GetHealthCheckTimeout()) * time.Nanosecond
		healthPeriod := normalizedTime / HealthCheckIntervalDivisor

		// todo(): add a way to stop the dispatcher
		// Increase work group to wait for the health status goroutine and spawn health check reporter
		wg.Add(1)
		go func() {
			// Ensure Done is called when the goroutine finishes
			defer wg.Done()

			// Report health status to the load balancer
			for {
				ctx, cancel := context.WithTimeout(context.Background(), normalizedTime)
				defer cancel()

				if err = client.ReportHealthStatus(ctx, &v1Consensus.HealthStatus{
					Service: "gale-node",
					Status:  uint32(v1Consensus.Serving),
					Message: "Serving requests goes brrrrr",
				}); err != nil {
					d.cfg.Logger.Errorf("failed to report health status: %v", err)
				}

				d.cfg.Logger.Debugf("reported health status to %s:%d", v.IP, v.Port)

				<-time.After(healthPeriod)
			}
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()
}

func (d *Dispatcher) Stop() {

}

/*

	var err error
	var wg sync.WaitGroup

	client, err := nodeNet.NewClient(cfg.Logger, address.IP, address.Port)
		if err != nil {
			cfg.Logger.Fatalf("failed to create client: %v", err)
		}

		ctxConfig, cancelConfig := context.WithTimeout(context.Background(), GetConfigTimeout)
		defer cancelConfig()

		executionCfg, err := client.GetConfig(ctxConfig)
		if err != nil {
			cfg.Logger.Fatalf("failed to get config: %v", err)
		}

		normalizedTime := time.Duration(executionCfg.HealthCheckTimeout) * time.Nanosecond
		// todo(): change this to a more accurate value
		healthPeriod := normalizedTime / 2

		// Increase work group to wait for the health status goroutine and spawn health check reporter
		wg.Add(1)
		go func() {
			// Ensure Done is called when the goroutine finishes
			defer wg.Done()

			// Report health status to the load balancer
			for {
				ctx, cancel := context.WithTimeout(context.Background(), normalizedTime)
				defer cancel()

				if err = client.ReportHealthStatus(ctx, &v1Consensus.HealthStatus{
					Service: "gale-node",
					Status:  uint32(v1Consensus.Serving),
					Message: "Serving requests goes brrrrr",
				}); err != nil {
					cfg.Logger.Errorf("failed to report health status: %v", err)
				}

				cfg.Logger.Debugf("reported health status to %s:%d", address.IP, address.Port)

				<-time.After(healthPeriod)
			}
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()
*/
