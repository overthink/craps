package main

import (
	"math/rand"
	"sync"
)

type PRNG struct {
	mutex  sync.Mutex
	source rand.Source
}

func (l *PRNG) Roll2() (int64, int64) {
	// we basically always need two rolls, so do both under same lock
	l.mutex.Lock()
	defer l.mutex.Unlock()
	a := l.source.Int63()%6 + 1
	b := l.source.Int63()%6 + 1
	return a, b
}

func NewPRNG(seed int64) PRNG {
	return PRNG{
		source: rand.NewSource(seed),
	}
}
