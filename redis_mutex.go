package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/fzzy/radix/redis"
	"sync"
	"time"
)

// ensure the RedisMutex follows the Locker interface
var _ = Locker(&RedisMutex{})

type RedisConfig struct {
	Servers []string
}

func (c *RedisConfig) Type() string {
	return "redis"
}

// Fields of a Mutex must not be changed after first use.
type RedisMutex struct {
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

// NewMutex returns a new Mutex on a named resource connected to the Redis instances at given addresses.
func NewRedisMutex(name string, servers []string, logger Logger) (*RedisMutex, error) {
	if len(servers) == 0 {
		return nil, errors.New("redis: servers are empty")
	}

	timeout := 500 * time.Millisecond

	nodes := []*redis.Client{}
	for _, server := range servers {
		node, err := redis.DialTimeout("tcp", server, timeout)
		if err != nil {
			logger.Debugf("Can't connect to %s: '%s'\n", server, err)
			continue
		}
		nodes = append(nodes, node)
	}

	if len(nodes) < 1 {
		return nil, ErrNoConnect
	}

	return &RedisMutex{
		Name:   name,
		Quorum: len(nodes)/2 + 1,
		Nodes:  nodes,
		logger: logger,
	}, nil
}

func (m *RedisMutex) SetExpiry(i time.Duration) {
	m.Expiry = i
}

func (m *RedisMutex) SetTries(i int) {
	m.Tries = i
}

func (m *RedisMutex) SetDelay(i time.Duration) {
	m.Delay = i
}

func (m *RedisMutex) SetFactor(i float64) {
	m.Factor = i
}

// Lock locks m.
// In case it returns an error on failure, you may retry to acquire the lock by calling this method again.
func (m *RedisMutex) Lock() error {
	m.nodem.Lock()
	defer m.nodem.Unlock()

	// get a random key
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return err
	}
	randomValue := base64.StdEncoding.EncodeToString(b)

	for i := 0; i < m.Tries; i++ {
		n := 0
		start := time.Now()

		for _, node := range m.Nodes {
			if node == nil {
				continue
			}

			reply := node.Cmd("set", m.Name, randomValue, "nx", "px", int(m.Expiry/time.Millisecond))
			if reply.Err != nil {
				m.logger.Debugf("during locking %s: %s\n", node.Conn.RemoteAddr(), reply.Err)
				continue
			}
			if reply.String() != "OK" {
				m.logger.Debugf("lock already taken on %s: '%s'\n", node.Conn.RemoteAddr(), reply.String())
				continue
			}
			m.logger.Debugf("lock aquired on %s\n", node.Conn.RemoteAddr())
			n += 1
		}

		until := time.Now().Add(m.Expiry - time.Now().Sub(start) - time.Duration(int64(float64(m.Expiry)*m.Factor)) + 2*time.Millisecond)
		if n >= m.Quorum && time.Now().Before(until) {
			m.value = randomValue
			m.until = until
			return nil
		}

		// lock failed, cleanup on nodes where the lock was acquired
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

		time.Sleep(m.Delay)
	}

	return ErrFailed
}

// Unlock unlocks m.
func (m *RedisMutex) Unlock() {
	m.nodem.Lock()
	defer m.nodem.Unlock()

	value := m.value
	// It is a run-time error if m is not locked on entry to Unlock.
	if value == "" {
		panic("unlock of unlocked mutex")
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
		m.logger.Debugf("unlocked %s\n", node.Conn.RemoteAddr())
	}
}
