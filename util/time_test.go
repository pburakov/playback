package util

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestJitter(t *testing.T) {
	rand.Seed(time.Now().Unix())

	j := Jitter(100)

	assert.True(t, j.Nanoseconds() < 100*1000000)
	assert.True(t, -100*1000000 < j.Nanoseconds())

	j = Jitter(100)

	assert.True(t, j.Nanoseconds() < 100*1000000)
	assert.True(t, -100*1000000 < j.Nanoseconds())

	j = Jitter(100)

	assert.True(t, j.Nanoseconds() < 100*1000000)
	assert.True(t, -100*1000000 < j.Nanoseconds())
}

func TestMSecToDuration(t *testing.T) {
	assert.Equal(t, 42*time.Millisecond, MSecToDuration(42))
	assert.Equal(t, -42*time.Millisecond, MSecToDuration(-42))
}
