package main

import (
	"testing"
	"time"
)

func TestNoAddresses(t *testing.T) {
	logger := &VerboseLog{}
	_, err := NewRedisMutex("keyname", []string{}, logger)
	expected := "redis: servers are empty"
	if err.Error() != expected {
		t.Errorf("got '%s', expected '%v'", err.Error(), expected)
	}
}

func TestCantConnect(t *testing.T) {
	logger := &SilentLog{}
	_, err := NewRedisMutex("keyname", []string{"10.0.0.0:6370"}, logger)
	if err == nil {
		t.Errorf("no connactable servers, should have thrown an error")
		return
	}

	if err != ErrNoConnect {
		t.Errorf("got '%s', expected '%v'", err, ErrNoConnect)
	}
}

func TestCanConnect(t *testing.T) {
	logger := &SilentLog{}
	m, err := NewRedisMutex("keyname", []string{"127.0.0.1:6379"}, logger)

	if err != nil {
		t.Errorf("should have been able to connect to 127.0.0.1:6370")
		return
	}

	expected := 1
	got := len(m.Nodes)
	if got != expected {
		t.Errorf("got '%s', expected '%s'", got, expected)
		return
	}

	expected = 1
	got = m.Quorum
	if got != expected {
		t.Errorf("got '%s', expected '%s'", got, expected)
		return
	}

	expectedName := "keyname"
	gotName := m.Name
	if got != expected {
		t.Errorf("got '%s', expected '%s'", expectedName, gotName)
		return
	}
}

func TestCanLock(t *testing.T) {
	logger := &SilentLog{}
	m, _ := NewRedisMutex("keyname", []string{"127.0.0.1:6379"}, logger)
	m.SetTries(1)
	m.SetExpiry(time.Millisecond)
	err := m.Lock()
	if err != nil {
		t.Errorf("locking failed, got '%s'", err)
		return
	}
	m.Unlock()
}

func BenchmarkLockUnlock(b *testing.B) {
	logger := &SilentLog{}
	m, _ := NewRedisMutex("keyname", []string{"127.0.0.1:6379"}, logger)
	m.SetTries(1)
	m.SetExpiry(time.Millisecond)
	for i := 0; i < b.N; i++ {
		m.Lock()
		m.Unlock()
	}
}
