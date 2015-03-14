package main

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"sync"
)

type Locker interface {
	Lock() error
	Unlock() error
}

// ensure the RedisMutex follows the Locker interface
var _ = Locker(&MysqlMutex{})

var ErrNoLock = errors.New("failed to acquire lock")

type MysqlMutex struct {
	Name    string // Key name
	Timeout int    // Duration for which the lock is valid, DefaultExpiry if 0
	db      *sql.DB
	nodem   sync.Mutex
}

// Lock locks a value
func (m *MysqlMutex) Lock() error {
	m.nodem.Lock()
	defer m.nodem.Unlock()

	sql := fmt.Sprintf("SELECT GET_LOCK('%s', %d);", m.Name, m.Timeout)
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
