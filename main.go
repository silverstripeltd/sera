package main

import (
	"errors"
	"fmt"
	"github.com/hjr265/redsync.go/redsync"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

func main() {
	var mutex *redsync.Mutex
	var err error

	var expiry time.Duration
	var command *exec.Cmd

	expiry, err = getExpiry()
	errorHandler(err)

	command, err = getCommand()
	errorHandler(err)

	addrs := []net.Addr{
		&net.TCPAddr{Port: 6379, IP: net.ParseIP("127.0.0.1")},
	}

	mutex, err = redsync.NewMutex("platformers-test-dev1", addrs)
	errorHandler(err)

	// Duration for which the lock is valid
	mutex.Expiry = expiry
	// Number of attempts to acquire lock before admitting failure, DefaultTries if 0
	mutex.Tries = 3
	// Delay between two attempts to acquire lock, DefaultDelay if 0
	mutex.Delay = time.Second

	err = mutex.Lock()
	errorHandler(err)

	var stdout io.ReadCloser
	stdout, err = command.StdoutPipe()
	errorHandler(err)
	go io.Copy(os.Stdout, stdout)

	var stderr io.ReadCloser
	stderr, err = command.StderrPipe()
	errorHandler(err)
	go io.Copy(os.Stderr, stderr)

	if err := command.Start(); err != nil {
		errorHandler(err)
	}

	if err := command.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0
			// This works on both Unix and Windows. Although package
			// syscall is generally platform dependent, WaitStatus is
			// defined for both Unix and Windows and in both cases has
			// an ExitStatus() method with the same signature.
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		errorHandler(err)
	}

	defer mutex.Unlock()

}

func getExpiry() (expiry time.Duration, err error) {
	args := os.Args[1:]
	if len(args) < 1 {
		err = errors.New("Usage: sera <expiry in sec> <command>")
		return
	}

	i, err := strconv.Atoi(args[0])
	errorHandler(err)

	expiry = time.Duration(i) * time.Second
	return
}

func getCommand() (cmd *exec.Cmd, err error) {
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

func errorHandler(err error) {
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(2)
	}
}
