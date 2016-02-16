package main

import (
	"os/exec"
	"testing"
	"time"
)

func TestCommandIsLogged(t *testing.T) {
	stat := &Statistic{}

	cmd := exec.Command("/some/command")
	cmd.Args = []string{"/some/command", "arg1", "arg2"}

	stat.SetCommand(cmd)

	expected := "/some/command arg1 arg2"
	actual := stat.Command
	if actual != expected {
		t.Errorf("Expected SetCommand() to set the command to '%s', got '%s'", expected, actual)
	}
}

func TestStart(t *testing.T) {
	stat := &Statistic{}

	startTime := time.Now()
	stat.Start()

	actual := stat.StartTime.Sub(startTime)
	if actual > time.Millisecond {
		t.Errorf("Expected Start() to be within one millisecond. got '%s'", actual)
	}
}

func TestStop(t *testing.T) {
	stat := &Statistic{}

	stat.Start()

	stopTime := time.Now()
	stat.Stop()

	actual := stat.EndTime.Sub(stopTime)
	if actual > time.Millisecond {
		t.Errorf("Expected Start() to be within one millisecond. got '%s'", actual)
	}
}

func TestStartStopsSetsDuration(t *testing.T) {
	stat := &Statistic{}

	if stat.Duration != 0 {
		t.Errorf("Duration should always be zero after initalization")
	}

	stat.Start()
	stat.Stop()

	if stat.Duration == 0 {
		t.Errorf("Duration should never be zero after Start()/Stop, got %s", stat.Duration)
	}
}

func TestStopWithoutStartWillNotCrash(t *testing.T) {
	stat := &Statistic{}
	stat.Stop()
	if stat.Duration != 0 {
		t.Errorf("Duration should be zero, got '%s'", stat.Duration)
	}
}

func TestGotLock(t *testing.T) {
	stat := &Statistic{}
	if !stat.LockAcquiredAt.IsZero() {
		t.Errorf("Expected that stat.LockAcquiredAt should be zero, got '%s'", stat.LockAcquiredAt)
	}

	stat.LockAcquired()
	if stat.LockAcquiredAt.IsZero() {
		t.Errorf("Expected that stat.LockAcquiredAt shouldn't be zero, got '%s'", stat.LockAcquiredAt)
	}
}

func TestStatsGetWrittenToFile(t *testing.T) {
	stat := &Statistic{}
	cmd := exec.Command("/some/command")
	cmd.Args = []string{"/some/command", "arg1", "arg2"}
	stat.SetCommand(cmd)
	stat.Start()
	stat.LockAcquired()
	stat.Stop()
	stat.Save("/tmp/stats.json")
	// @todo add test
	// @todo set the command exit code in the statistics
	// @todo ensure that the file is too big
	// @todo investigate file locks and backup with jitter on saving
	// @todo add the stat file location to the config file
}
