package nodenetwork

// Dispatcher contains the logic that determines how and when to dispatch messages to the load balancers. Dispatcher
// initializes N connections to N load balancers to todo()
type Dispatcher struct {
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}
