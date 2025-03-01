package common

type AddrKey struct {
	IP   uint32
	Port uint16
	Pad  uint16 // Padding for memory alignment (must match C struct)
}
