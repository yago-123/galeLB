package consensus

type ServiceStatus uint32

const (
	Serving ServiceStatus = iota
	NotServing
	ShuttingDown
)

func StatusString(s ServiceStatus) string {
	switch s {
	case Serving:
		return "Serving"
	case NotServing:
		return "Not Serving"
	case ShuttingDown:
		return "Shutting Down"
	default:
		return "Unknown"
	}
}
