package main

import (
	"flag"
	"os"
	"testing"
	"time"
)

func TestTimeoutArgWorks(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"sera", "10", "ls"}
	flag.CommandLine.Parse(os.Args[1:])

	expected := time.Second * 10
	var actual time.Duration
	timeoutArg(&actual)

	if actual != expected {
		t.Errorf("got '%v', expected '%v'", actual, expected)
	}
}

func TestTimeoutArgFailsOnNonInt(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"sera", "fail", "ls"}
	flag.CommandLine.Parse(os.Args[1:])

	expected := time.Duration(0)
	var actual time.Duration
	err := timeoutArg(&actual)

	if err == nil {
		t.Errorf("expected error")
	}

	if actual != expected {
		t.Errorf("got '%v', expected '%v'", actual, expected)
	}
}

func TestMd5Generation(t *testing.T) {
	input := "ls -ltra"
	expected := "5373fa7de6f7dd54052abf86b3a53da4"
	actual := md5Hash(input)
	if actual != expected {
		t.Errorf("got '%s', expected '%v'", actual, expected)
	}
}
