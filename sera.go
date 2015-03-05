package main

// sērus m (feminine sēra, neuter sērum); first/second declension
//
// late, too late
// slow, tardy
import (
	"errors"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
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

	var err error

	logger := &VerboseLog{}

	keyExpiry, err := Expiry()
	if err != nil {
		logger.Printf("%s\n", err)
		return 2
	}

	// the key name used for locking, should be unique per environment and command
	keyName := "platformers-test-dev1"

	logger.Printf("expiry: %s, name: %s\n", keyExpiry, keyName)

	addrs := []net.Addr{
		&net.TCPAddr{Port: 6379, IP: net.ParseIP("127.0.0.1")},
		&net.TCPAddr{Port: 6380, IP: net.ParseIP("127.0.0.1")},
		&net.TCPAddr{Port: 6381, IP: net.ParseIP("127.0.0.1")},
	}

	mutex, err := NewMutex(keyName, addrs, logger)

	if err != nil {
		if err == ErrNoConnect {
			logger.Debugf("No quorum servers alive, run the command anyway..\n")
			return 1
		} else {
			logger.Printf("%s\n", err)
			return 2
		}
	}

	// Duration for which the lock is valid
	mutex.Expiry = keyExpiry
	// Number of attempts to acquire lock before admitting failure, DefaultTries if 0
	mutex.Tries = 16
	// Delay between two attempts to acquire lock
	mutex.Delay = keyExpiry / time.Duration(mutex.Tries)

	err = mutex.Lock()
	if err != nil {
		logger.Printf("%s\n", err)
		return 2
	}
	defer mutex.Unlock()

	cmd, err := Command()
	if err != nil {
		logger.Printf("%s\n", err)
		return 2
	}
	cmd.Args = os.Args[2:]

	var stdout io.ReadCloser
	stdout, err = cmd.StdoutPipe()

	go io.Copy(os.Stdout, stdout)

	var stderr io.ReadCloser
	stderr, err = cmd.StderrPipe()

	go io.Copy(os.Stderr, stderr)

	if err := cmd.Start(); err != nil {
        logger.Printf("arg, Fail!")
		return 2
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0
			// This works on both Unix and Windows. Although package
			// syscall is generally platform dependent, WaitStatus is
			// defined for both Unix and Windows and in both cases has
			// an ExitStatus() method with the same signature.
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus()
			}
		}
        logger.Printf("Command failed with %s\n", err)

	}

	return 0
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
