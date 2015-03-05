package main

import (
	"fmt"
	"os"
	"time"
)

type Logger interface {
	Printf(format string, a ...interface{}) (n int, err error)
	Debugf(format string, a ...interface{}) (n int, err error)
}

type VerboseLog struct{}

func (log *VerboseLog) Printf(format string, a ...interface{}) (n int, err error) {
	format = time.Now().Format(time.RFC3339) + " " + format
	return fmt.Fprintf(os.Stdout, format, a...)
}

func (log *VerboseLog) Debugf(format string, a ...interface{}) (n int, err error) {
	format = time.Now().Format(time.RFC3339) + " " + format
	return fmt.Fprintf(os.Stdout, format, a...)
}

type SilentLog struct {
	Prints []string
	Debugs []string
}

func (log *SilentLog) Printf(format string, a ...interface{}) (n int, err error) {
	log.Prints = append(log.Prints, fmt.Sprintf(format, a...))
	return 0, nil
}

func (log *SilentLog) Debugf(format string, a ...interface{}) (n int, err error) {
	log.Debugs = append(log.Debugs, fmt.Sprintf(format, a...))
	return 0, nil
}
