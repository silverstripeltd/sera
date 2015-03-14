package main

import (
    "testing"
)

func TestMd5Generation(t *testing.T) {
    input :=  "ls -ltra"
    expected := "5373fa7de6f7dd54052abf86b3a53da4"
    actual := md5Hash(input)
    if actual != expected {
        t.Errorf("got '%s', expected '%v'", actual, expected)
    }
}