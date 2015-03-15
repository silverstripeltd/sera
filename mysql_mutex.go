package main

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"sync"
	"time"
)

// Interface for objects that attempt to get locks, but are not allowed
// to block forever. The implementors are expected to timeout at some point.
type LockerWithTimeout interface {
	Lock() error
	Unlock() error
	SetDuration(time.Duration)
}

// ensure the MysqlMutex follows the Locker interface
var _ = LockerWithTimeout(&MysqlMutex{})

var ErrNoLock = errors.New("failed to acquire lock")

type MysqlMutex struct {
	Name     string        // Key name
	Duration time.Duration // Duration for which the lock is valid, DefaultExpiry if 0
	db       *sql.DB
	nodem    sync.Mutex
}

// Set the timeout duration.
func (m *MysqlMutex) SetDuration(duration time.Duration) {
	m.Duration = duration
}

// Lock locks a value
func (m *MysqlMutex) Lock() error {
	// Make the process of getting the mysql lock atomic by wrapping it in a mutex.
	m.nodem.Lock()
	defer m.nodem.Unlock()

	sql := fmt.Sprintf("SELECT GET_LOCK('%s', %d);", m.Name, int(m.Duration.Seconds()))
	rows, err := m.db.Query(sql)
	if err != nil {
		return err
	}
	defer rows.Close()

	var value int
	for rows.Next() {
		if err := rows.Scan(&value); err != nil {
			return err
		}
	}
	if value != 1 {
		return ErrNoLock
	}
	return nil
}

// Unlock unlocks m.
func (m *MysqlMutex) Unlock() error {
	m.nodem.Lock()
	defer m.nodem.Unlock()

	sql := fmt.Sprintf("SELECT RELEASE_LOCK('%s');", m.Name)
	rows, err := m.db.Query(sql)
	if err != nil {
		return err
	}
	rows.Close()
	return nil
}
