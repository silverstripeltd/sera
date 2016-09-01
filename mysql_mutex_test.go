package main

import (
	"errors"
	"testing"
)

type dbMock struct {
	mockError bool
	mockValue int
}

func (d *dbMock) Query(query string, args ...interface{}) (QueryableResponse, error) {
	if d.mockError {
		return nil, errors.New("Terrible crash")
	}
	return &respMock{
		mockValue:     d.mockValue,
		rowsAvailable: 1,
	}, nil
}

type respMock struct {
	mockValue     int
	rowsAvailable int
}

func (r *respMock) Close() error {
	return nil
}
func (r *respMock) Next() bool {
	haveSome := r.rowsAvailable > 0
	r.rowsAvailable--
	return haveSome
}
func (r *respMock) Scan(dest ...interface{}) error {
	*(dest[0].(*int)) = r.mockValue
	return nil
}

func TestLockDBError(t *testing.T) {
	db := &dbMock{
		mockError: true,
	}
	mutex := NewMysqlMutex(db, "foo", 0)
	err := mutex.Lock()
	if err == nil {
		t.Errorf("Expected error on query error.\n")
	}
}

func TestLockAcquired(t *testing.T) {
	db := &dbMock{
		mockError: false,
		mockValue: 1,
	}
	mutex := NewMysqlMutex(db, "foo", 0)
	err := mutex.Lock()
	if err != nil {
		t.Errorf("Error '%s' was unexpected", err)
	}
}

func TestLockNotAcquired(t *testing.T) {
	db := &dbMock{
		mockError: false,
		mockValue: 0,
	}
	mutex := NewMysqlMutex(db, "foo", 0)
	err := mutex.Lock()
	if err == nil {
		t.Errorf("Expected error on query error.\n")
	}
}
