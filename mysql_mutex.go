package main

// create database sera;
// CREATE USER 'sera'@'localhost' IDENTIFIED BY 'secret';
// GRANT ALL PRIVILEGES ON sera . * TO 'sera'@'localhost';
// CREATE TABLE mutex ( name VARCHAR(255) NOT NULL, expiry TIMESTAMP NOT NULL );
// "set", m.Name, randomValue, "nx", "px", int(expiry/time.Millisecond

import (
	//	"crypto/rand"
	//	"encoding/base64"
	//	"crypto/rand"
	"database/sql"
	//	"encoding/base64"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"sync"
	"time"
)

// ensure the RedisMutex follows the Locker interface
var _ = Locker(&RedisMutex{})

type MysqlConfig struct {
	Servers []string
}

func (c *MysqlConfig) Type() string {
	return "mysql"
}

func (c *MysqlConfig) Backends() []string {
	return c.Servers
}

// Fields of a Mutex must not be changed after first use.
type MysqlMutex struct {
	Name   string        // Resouce name
	Expiry time.Duration // Duration for which the lock is valid, DefaultExpiry if 0

	Tries int           // Number of attempts to acquire lock before admitting failure, DefaultTries if 0
	Delay time.Duration // Delay between two attempts to acquire lock, DefaultDelay if 0

    Factor float64 // Drift factor, DefaultFactor if 0

    Quorum int
	value  string
	until  time.Time
	logger Logger

	Nodes []*sql.DB

	nodem sync.Mutex
}

// NewMutex returns a new Mutex on a named resource connected to the Redis instances at given addresses.
func NewMysqlMutex(name string, servers []string, logger Logger) (*MysqlMutex, error) {
	if len(servers) == 0 {
		return nil, errors.New("mysql: servers are empty")
	}

	nodes := []*sql.DB{}

	for _, server := range servers {
		db, err := sql.Open("mysql", server)
		if err != nil {
			panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
		}

		err = db.Ping()
		if err != nil {
			logger.Printf("Cant connect to %s: '%s'\n", server, err)
			continue
		}

		nodes = append(nodes, db)
	}

	return &MysqlMutex{
		Name:   name,
		Nodes:  nodes,
        Quorum: len(nodes)/2 + 1,
		logger: logger,
	}, nil
}

// Lock locks m.
// In case it returns an error on failure, you may retry to acquire the lock by calling this method again.
func (m *MysqlMutex) Lock() error {
	m.nodem.Lock()
	defer m.nodem.Unlock()

	// get a random key
	//	b := make([]byte, 16)
	//	_, err := rand.Read(b)
	//	if err != nil {
	//		return err
	//	}
	//	randomValue := base64.StdEncoding.EncodeToString(b)

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

    factor := m.Factor
    if factor == 0 {
        factor = DefaultFactor
    }

	for i := 0; i < retries; i++ {
		n := 0
		//		start := time.Now()

		for _, node := range m.Nodes {
			if node == nil {
				continue
			}

			sql := fmt.Sprintf("SELECT GET_LOCK('%s', %d);", m.Name, int(m.Expiry.Seconds()))
            rows, err := node.Query(sql)
			if err != nil {
				m.logger.Printf("%s\n", err)
			}
            defer rows.Close()

            var value int
            for rows.Next() {
                err := rows.Scan(&value)
                if err != nil {
                    m.logger.Debugf("During locking %s\n", err)
                }
            }

            if value != 1 {
                m.logger.Debugf("Lock already taken\n",)
                continue
            }

            m.logger.Debugf("Lock aquired\n",)
            n += 1
		}

        // lock aquired
        if n >= m.Quorum {
            return nil
        }

		time.Sleep(delay)
	}

	return ErrFailed
}

// Unlock unlocks m.
// It is a run-time error if m is not locked on entry to Unlock.
func (m *MysqlMutex) Unlock() {
	m.nodem.Lock()
	defer m.nodem.Unlock()

    for _, node := range m.Nodes {
        if node == nil {
            continue
        }

        sql := fmt.Sprintf("SELECT RELEASE_LOCK('%s');", m.Name)
        rows, err := node.Query(sql)
        if err != nil {
            m.logger.Printf("%s\n", err)
        }
        defer rows.Close()

        m.logger.Debugf("Unlocked\n")
    }

}

func (m *MysqlMutex) Value() string {
	return m.value
}

func (m *MysqlMutex) Until() time.Time {
	return m.until
}
