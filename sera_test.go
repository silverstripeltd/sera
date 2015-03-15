package main

import (
    "testing"
    "os"
)

func TestTimeoutArgWorks(t *testing.T) {
    oldArgs := os.Args
    defer func() { os.Args = oldArgs }()
    os.Args = []string{"sera", "10", "ls"}

    expected := 10
    var actual int
    timeoutArg(&actual)

    if actual != expected {
        t.Errorf("got '%v', expected '%v'", actual, expected)
    }
}

func TestTimeoutArgFailsOnNonInt(t *testing.T) {
    oldArgs := os.Args
    defer func() { os.Args = oldArgs }()
    os.Args = []string{"sera", "fail", "ls"}

    expected := 0
    var actual int
    err := timeoutArg(&actual)

    if err == nil {
        t.Errorf("expected error")
    }

    if actual != expected {
        t.Errorf("got '%v', expected '%v'", actual, expected)
    }
}

func TestMd5Generation(t *testing.T) {
    input :=  "ls -ltra"
    expected := "5373fa7de6f7dd54052abf86b3a53da4"
    actual := md5Hash(input)
    if actual != expected {
        t.Errorf("got '%s', expected '%v'", actual, expected)
    }
}