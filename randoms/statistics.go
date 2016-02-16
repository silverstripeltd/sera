package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Statistics struct {
	Statistics []*Statistic
}

type Statistic struct {
	StartTime      time.Time
	EndTime        time.Time
	Duration       time.Duration
	Command        string
	Timeout        time.Duration
	LockAcquiredAt time.Time
	ExitCode       int
}

func (s *Statistic) SetCommand(c *exec.Cmd) {
	s.Command = fmt.Sprintf("%s", strings.Join(c.Args, " "))
}

func (s *Statistic) Start() {
	s.StartTime = time.Now()
}

func (s *Statistic) Stop() {
	s.EndTime = time.Now()
	if s.StartTime.IsZero() {
		return
	}
	s.Duration = s.EndTime.Sub(s.StartTime)
}

func (s *Statistic) LockAcquired() {
	s.LockAcquiredAt = time.Now()
}

func (s *Statistic) SetTimeout(t time.Duration) {
	s.Timeout = t
}

func (s *Statistic) Save(filepath string) {

	var fileData []byte
	var stats Statistics

	if _, err := os.Stat(filepath); !os.IsNotExist(err) {
		fileData, err = ioutil.ReadFile(filepath)
		if err != nil {
			log(fmt.Sprintf("error opening file '%s': %v", filepath, err))
			return
		}
		// Unmarshal the json string into the stats struct
		if err := json.Unmarshal(fileData, &stats); err != nil {
			log(fmt.Sprintf("error json.Unmarshal file '%s': %v", filepath, err))
			return
		}
	}

	stats.Statistics = append(stats.Statistics, s)
	fileData, err := json.Marshal(stats)
	if err != nil {
		log("error when encoding statistics to JSON")
		return
	}
	if err = ioutil.WriteFile(filepath, fileData, 0666); err != nil {
		log(fmt.Sprintf("error while saving file '%s': %v", filepath, err))
		return
	}
}
