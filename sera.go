package main

// sērus m (feminine sēra, neuter sērum); first/second declension
//
// late, too late
// slow, tardy
import (
	"errors"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	// Call realMain instead of doing the work here so we can use
	// `defer` statements within the function and have them work properly.
	// (defers aren't called with os.Exit)
	os.Exit(realMain())
}

func realMain() int {

	logger := &VerboseLog{}

	conf, err := NewConfig("conf.json")
	if err != nil {
		logger.Printf("%s\n", err)
		return 2
	}

	keyExpiry, err := Expiry()
	if err != nil {
		logger.Printf("%s\n", err)
		return 2
	}

	keyName := strings.Join(os.Args[2:], " ")

	logger.Printf("expiry: %s, name: '%s'\n", keyExpiry, keyName)

	mutex, err := MutexFactory(conf, keyName, keyExpiry, logger)

	if err != nil {
		if err == ErrNoConnect {
			logger.Debugf("No quorum servers alive, run the command anyway..\n")
			return 1
		} else {
			logger.Printf("%s\n", err)
			return 2
		}
	}

	err = mutex.Lock()
	if err != nil {
		logger.Printf("%s\n", err)
		return 2
	}
	defer mutex.Unlock()

	cmd, err := Command()
	if err != nil {
		logger.Printf("%s\n", err)
		return 127
	}
	cmd.Args = os.Args[2:]

	err = PipeCommandOutput(cmd)
	if err != nil {
		logger.Printf("%s\n", err)
		return 2
	}

	exitStatus, err := RunCommand(cmd)
	if err != nil {
		logger.Printf("%s\n", err)
		return exitStatus
	}

	return exitStatus
}

func RunCommand(cmd *exec.Cmd) (existatus int, err error) {
	// generic failure
	if err := cmd.Start(); err != nil {
		return 2, err
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0
			// This works on both Unix and Windows. Although package
			// syscall is generally platform dependent, WaitStatus is
			// defined for both Unix and Windows and in both cases has
			// an ExitStatus() method with the same signature.
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus(), nil
			}
		}
		return 2, err
	}
	return 0, nil
}

func MutexFactory(conf Config, keyName string, expiry time.Duration, logger Logger) (Locker, error) {

	if conf.Type() == "redis" {
		mutex, err := NewRedisMutex(keyName, conf.Backends(), logger)
		// Duration for which the lock is valid
		mutex.Expiry = expiry
		// Number of attempts to acquire lock before admitting failure, DefaultTries if 0
		mutex.Tries = 16
		// Delay between two attempts to acquire lock
		mutex.Delay = mutex.Expiry / time.Duration(mutex.Tries)
		return mutex, err
	}

	if conf.Type() == "mysql" {
		mutex, err := NewMysqlMutex(keyName, conf.Backends(), logger)
		// Duration for which the lock is valid
		mutex.Expiry = expiry
		// Number of attempts to acquire lock before admitting failure, DefaultTries if 0
		//        mutex.Tries = 16
		// Delay between two attempts to acquire lock
		//        mutex.Delay = mutex.Expiry / time.Duration(mutex.Tries)
		return mutex, err
	}

	return nil, errors.New("Could not find Locker for backend type '" + conf.Type() + "'")
}

// get the commands stdout and stderr
func PipeCommandOutput(cmd *exec.Cmd) (err error) {

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	// get the commands stdout and stderr
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// As soon the command prints to its stdout/stderr, print to the "real" stdout/stderr
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	return nil
}

func Expiry() (expiry time.Duration, err error) {
	args := os.Args[1:]
	if len(args) < 1 {
		err = errors.New("Usage: sera <expiry in sec> <command>")
		return
	}

	seconds, err := strconv.Atoi(args[0])
	if err != nil {
		return 0, err
	}
	expiry = time.Duration(seconds) * time.Second
	return
}

func Command() (cmd *exec.Cmd, err error) {
	args := os.Args[1:]
	if len(args) < 2 {
		err = errors.New("Usage: sera <expiry in sec> <command>")
		return
	}

	var path string
	path, err = exec.LookPath(args[1])

	cmd = exec.Command(path)
	cmd.Args = os.Args[2:]

	return
}
