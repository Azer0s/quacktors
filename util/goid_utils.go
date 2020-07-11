package util

import (
	"bytes"
	"runtime"
	"strconv"
)

// GetGoid returns the goroutine ID of a goroutine (this is sort of a hack)
func GetGoid() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}
