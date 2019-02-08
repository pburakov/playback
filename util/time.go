package util

import (
	"math/rand"
	"time"
)

func DefaultTimestamp() time.Time {
	return time.Unix(0, 0)
}

func Jitter(maxJitterMSec int) time.Duration {
	return time.Duration(float64(maxJitterMSec) * ((rand.Float64() - 0.5) * 2) * 1000000)
}

func MSecToDuration(ms int) time.Duration {
	return time.Duration(ms * 1000000)
}
