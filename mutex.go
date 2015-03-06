package main

// A Mutex is a mutual exclusion lock.

import (
	"errors"
	"time"
)

const (
	DefaultExpiry = 8 * time.Second
	DefaultTries  = 16
	DefaultDelay  = 512 * time.Millisecond
	DefaultFactor = 0.01
)

var (
	ErrFailed    = errors.New("failed to acquire lock")
	ErrNoConnect = errors.New("Failed to connect to any redis server")
)

type Locker interface {
	Lock() error
	Unlock()
	Value() string
	Until() time.Time
}
