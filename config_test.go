package main

import (
	"testing"
)

func TestImport(t *testing.T) {

	conf, err := NewConfig("conf.json")
	if err != nil {
		t.Errorf("%s\n", err)
		return
	}

	if conf.Type() != "redis" {
		t.Errorf("config backend is not 'redis', got '%s'", conf.Type())
	}
}
