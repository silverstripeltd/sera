package main

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Queryable interface {
	Query(query string, args ...interface{}) (QueryableResponse, error)
}
type QueryableResponse interface {
	Close() error
	Next() bool
	Scan(dest ...interface{}) error
}

// sql.DB must be wrapped because the Query method uses a concrete return type (sql.Rows)
// which does not fit with Queryable interface and prevents us from mocking it.
type MysqlConnection struct {
	DB *sql.DB
}

func (c *MysqlConnection) Query(query string, args ...interface{}) (QueryableResponse, error) {
	return c.DB.Query(query, args...)
}

type ErrLockTimeout string

func (e ErrLockTimeout) Error() string {
	return string(e)
}

func NewMysqlMutex(q Queryable, keyName string, timeout time.Duration) *MysqlMutex {
	return &MysqlMutex{
		Name:    md5Hash(keyName),
		db:      q,
		Timeout: timeout,
	}
}

type MysqlMutex struct {
	Name    string        // Key name
	Timeout time.Duration // Duration for which the lock is valid, DefaultExpiry if 0
	db      Queryable
	nodem   sync.Mutex
}

// Lock uses the MySQL GET_LOCK() function to ensure that only one caller at time can hold
// a lock with the same MysqlMutex.Name. It blocks (wait) until MysqlMutex.Timeout value
// have passed and then returns an ErrNoLock error.
func (m *MysqlMutex) Lock() error {
	// Make the process of getting the mysql lock atomic by wrapping it in a mutex in
	// case multiple processes are trying to call this process.
	m.nodem.Lock()
	defer m.nodem.Unlock()

	sql := fmt.Sprintf("SELECT GET_LOCK('%s', %d);", m.Name, int(m.Timeout.Seconds()))
	rows, err := m.db.Query(sql)
	if err != nil {
		return err
	}
	defer rows.Close()

	// the result of the SELECT GET_LOCK() query should be a single row of 1 (true) for
	// successful locking or 0 (false) for a lock that is already taken.
	var value int
	for rows.Next() {
		if err := rows.Scan(&value); err != nil {
			return err
		}
	}
	if value != 1 {
		return ErrLockTimeout("Timeout out trying to acquire the lock.")
	}
	return nil
}

// Unlock releases the lock
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
