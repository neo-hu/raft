package raft

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func afterBetween(min time.Duration, max time.Duration) time.Duration {
	d, delta := min, max-min
	if delta > 0 {
		d += time.Duration(rand.Int63n(int64(delta)))
	}
	return d
}
