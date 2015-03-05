package main

import (
	"encoding/json"
	"errors"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
)

type Config interface {
	Type() string
    Backends() []string
}

type RedisConfig struct {
	Servers []string
}

func (c *RedisConfig) Type() string {
	return "redis"
}

func (c *RedisConfig) Backends() []string {
    return c.Servers
}

func NewConfig(path string) (Config, error) {

	jsonStr, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal(jsonStr, &data)
	if err != nil {
		return nil, err
	}

	switch data["backend"] {
	case "redis":
		config := &RedisConfig{}
		err := mapstructure.Decode(data, &config)
		return config, err
	}

	return nil, errors.New("Can't find config backend '" + data["backend"].(string) + "' from " + path)
}
