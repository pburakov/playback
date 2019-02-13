package util

import (
	"math/rand"
	"time"
)

// DefaultTimestamp returns zero unix time timestamp.
func DefaultTimestamp() time.Time {
	return time.Unix(0, 0)
}

// Jitter calculates random jitter duration value that deviates within given max
// jitter constraint, as a negative and positive offset from zero.
func Jitter(maxJitterMSec int) time.Duration {
	return time.Duration(float64(maxJitterMSec) * ((rand.Float64() - 0.5) * 2) * 1000000)
}

// MSecToDuration converts integer millisecond value to duration value.
func MSecToDuration(ms int) time.Duration {
	return time.Duration(ms * 1000000)
}
