package routing

import "hash/crc32"

type Hasher func([]byte) uint32

func Crc32Hasher(key []byte) uint32 {
	return crc32.ChecksumIEEE(key)
}
