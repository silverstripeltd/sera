package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/fzzy/radix/redis"
	"net"
	"sync"
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
}

// A Mutex is a mutual exclusion lock.
//
// Fields of a Mutex must not be changed after first use.
type Mutex struct {
	Name   string        // Resouce name
	Expiry time.Duration // Duration for which the lock is valid, DefaultExpiry if 0

	Tries int           // Number of attempts to acquire lock before admitting failure, DefaultTries if 0
	Delay time.Duration // Delay between two attempts to acquire lock, DefaultDelay if 0

	Factor float64 // Drift factor, DefaultFactor if 0

	Quorum int // Quorum for the lock, set to len(addrs)/2+1 by NewMutex()

	value string
	until time.Time

	Nodes []*redis.Client
	nodem sync.Mutex

	logger Logger
}

var _ = Locker(&Mutex{})

// NewMutex returns a new Mutex on a named resource connected to the Redis instances at given addresses.
func NewMutex(name string, addrs []net.Addr, ui Logger) (*Mutex, error) {
	if len(addrs) == 0 {
		return nil, errors.New("redis: addrs is empty")
	}

	timeout := 500 * time.Millisecond

	nodes := []*redis.Client{}
	for _, addr := range addrs {
		node, err := redis.DialTimeout(addr.Network(), addr.String(), timeout)
		if err == nil {
			nodes = append(nodes, node)
		}
	}

	if len(nodes) < 1 {
		return nil, ErrNoConnect
	}

	return &Mutex{
		Name:   name,
		Quorum: len(nodes)/2 + 1,
		Nodes:  nodes,
		logger: ui,
	}, nil
}

// Lock locks m.
// In case it returns an error on failure, you may retry to acquire the lock by calling this method again.
func (m *Mutex) Lock() error {
	m.nodem.Lock()
	defer m.nodem.Unlock()

	// get a random key
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return err
	}
	randomValue := base64.StdEncoding.EncodeToString(b)

	expiry := m.Expiry
	if expiry == 0 {
		expiry = DefaultExpiry
	}

	retries := m.Tries
	if retries == 0 {
		retries = DefaultTries
	}

	delay := m.Delay
	if delay == 0 {
		delay = DefaultDelay
	}

	for i := 0; i < retries; i++ {
		n := 0
		start := time.Now()

		for _, node := range m.Nodes {
			if node == nil {
				continue
			}

			reply := node.Cmd("set", m.Name, randomValue, "nx", "px", int(expiry/time.Millisecond))
			if reply.Err != nil {
				m.logger.Debugf("During locking %s: %s\n", node.Conn.RemoteAddr(), err)
				continue
			}
			if reply.String() != "OK" {
				m.logger.Debugf("Lock already taken on %s\n", node.Conn.RemoteAddr())
				continue
			}
			m.logger.Debugf("Lock aquired on %s\n", node.Conn.RemoteAddr())
			n += 1
		}

		factor := m.Factor
		if factor == 0 {
			factor = DefaultFactor
		}

		until := time.Now().Add(expiry - time.Now().Sub(start) - time.Duration(int64(float64(expiry)*factor)) + 2*time.Millisecond)
		if n >= m.Quorum && time.Now().Before(until) {
			m.value = randomValue
			m.until = until
			return nil
		}

		// lock failed, cleanup on nodes where the lock was aquired
		for _, node := range m.Nodes {
			if node == nil {
				continue
			}

			node.Cmd("eval", `
                if redis.call("get", KEYS[1]) == ARGV[1] then
                    return redis.call("del", KEYS[1])
                else
                    return 0
                end
            `, 1, m.Name, randomValue)
		}

		time.Sleep(delay)
	}

	return ErrFailed
}

// Unlock unlocks m.
// It is a run-time error if m is not locked on entry to Unlock.
func (m *Mutex) Unlock() {
	m.nodem.Lock()
	defer m.nodem.Unlock()

	value := m.value
	if value == "" {
		panic("redis: unlock of unlocked mutex")
	}

	m.value = ""
	m.until = time.Unix(0, 0)

	for _, node := range m.Nodes {
		if node == nil {
			continue
		}

		node.Cmd("eval", `
			if redis.call("get", KEYS[1]) == ARGV[1] then
			    return redis.call("del", KEYS[1])
			else
			    return 0
			end
		`, 1, m.Name, value)
		m.logger.Debugf("Unlocked %s\n", node.Conn.RemoteAddr())
	}
}

func (m *Mutex) Value() string {
	return m.value
}

func (m *Mutex) Until() time.Time {
	return m.until
}
