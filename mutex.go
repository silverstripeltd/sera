package main

// A Mutex is a mutual exclusion lock.

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
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
	ErrNoConnect = errors.New("Failed to connect to any servers")
)

type Locker interface {
	Lock() error
	Unlock()
	SetExpiry(i time.Duration)
	SetTries(i int)
	SetDelay(i time.Duration)
	SetFactor(i float64)
}

func NewMutex(path string, name string, logger Logger) (mutex Locker, err error) {

	// read config file
	jsonStr, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// just map the config to a generic interface
	var data map[string]interface{}
	err = json.Unmarshal(jsonStr, &data)
	if err != nil {
		return nil, err
	}

	switch data["backend"] {
	default:
		err = fmt.Errorf("Can't find config backend '%s' from %s", data["backend"].(string), path)
	case "redis":
		config := &RedisConfig{}
		if err := mapstructure.Decode(data, &config); err != nil {
			return nil, err
		}
		mutex, err = NewRedisMutex(name, config.Servers, logger)

	case "mysql":
		config := &MysqlConfig{}
		if err := mapstructure.Decode(data, &config); err != nil {
			return nil, err
		}
		mutex, err = NewMysqlMutex(name, config.Servers, logger)
	}


	if err == nil {
		mutex.SetDelay(DefaultDelay)
		mutex.SetExpiry(DefaultExpiry)
		mutex.SetTries(DefaultTries)
		mutex.SetFactor(DefaultFactor)
	}

	return mutex, err
}
