package consensus

type ServiceStatus uint32

const (
	Serving ServiceStatus = iota
	NotServing
)
