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
var _ = Locker(&MysqlMutex{})

type MysqlConfig struct {
	Servers []string
}

func (c *MysqlConfig) Type() string {
	return "mysql"
}

// Fields of a Mutex must not be changed after first use.
type MysqlMutex struct {
	Name   string        // Key name
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
	if len(servers) < 1 {
		return nil, errors.New("mysql: servers are empty")
	}

	nodes := []*sql.DB{}
	for _, server := range servers {
		db, err := sql.Open("mysql", server)
		if err != nil {
			panic(err.Error())
		}
		err = db.Ping()
		if err != nil {
			logger.Debugf("Can't connect to %s: '%s'\n", server, err)
			continue
		}
		nodes = append(nodes, db)
	}

	if len(nodes) < 1 {
		return nil, ErrNoConnect
	}

	return &MysqlMutex{
		Name:   name,
		Nodes:  nodes,
		Quorum: len(nodes)/2 + 1,
		logger: logger,
	}, nil
}

func (m *MysqlMutex) SetExpiry(i time.Duration) {
	m.Expiry = i
}

func (m *MysqlMutex) SetTries(i int) {
	m.Tries = i
}

func (m *MysqlMutex) SetDelay(i time.Duration) {
	m.Delay = i
}

func (m *MysqlMutex) SetFactor(i float64) {
	m.Factor = i
}

// Lock locks m.
// In case it returns an error on failure, you may retry to acquire the lock by calling this method again.
func (m *MysqlMutex) Lock() error {
	m.nodem.Lock()
	defer m.nodem.Unlock()

	for i := 0; i < m.Tries; i++ {
		numLockAquired := 0
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
				m.logger.Debugf("Lock already taken\n")
				continue
			}

			m.logger.Debugf("Lock aquired\n")
			numLockAquired += 1
		}

		// lock aquired
		if numLockAquired >= m.Quorum {
			return nil
		}

		m.logger.Debugf("Sleep %s\n", m.Delay)
		time.Sleep(m.Delay)
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
