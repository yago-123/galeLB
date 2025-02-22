package consensus

type ServiceStatus uint

const (
	Serving ServiceStatus = iota
	NotServing
)
