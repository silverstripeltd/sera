package main

// sera
//
// from latin: late, too late, slow

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log/syslog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var ErrUsage = errors.New("Usage: sera <wait-time-in-seconds> <command>")

var (
	conf      Config
	logger, _ = syslog.New(syslog.LOG_NOTICE, "sera")
)

// Define exit codes above the typical Linux exit code space, so we can
// reflect the exit values of the actual underlying program.
const (
	ExitSuccess          = 0
	ExitBadConfigFile    = 200
	ExitBadArguments     = 201
	ExitLockFailed       = 202
	ExitCommandRunFailed = 203
)

// Represents the json config for sera
type Config struct {
	Server  string `json:"server"`
	Syslog  bool   `json:"syslog"`
	Verbose bool   `json:"verbose"`
}

func main() {
	// Call realMain instead of doing the work here so we can use
	// `defer` statements within the function and have them work properly.
	// (defers aren't called with os.Exit)
	exitStatus, err := realMain()

	if exitStatus == ExitBadArguments {
		fmt.Println(ErrUsage)
		os.Exit(ExitBadArguments)
	}

	if err != nil {
		log(err)
	}

	os.Exit(exitStatus)
}

func realMain() (int, error) {

	// read config file
	jsonStr, err := ioutil.ReadFile("/etc/sera.json")
	if err != nil {
		return ExitBadConfigFile, err
	}
	// Unmarshal the json string into the config struct
	if err = json.Unmarshal(jsonStr, &conf); err != nil {
		return ExitBadConfigFile, err
	}

	// Ensure we got all of the required arguments
	if len(os.Args) < 3 {
		return ExitBadArguments, ErrUsage
	}

	// get the name of the key to use as a lock, in this case the command
	keyName := strings.Join(os.Args[2:], " ")

	// get the timeout for how long we will for the lock to be available
	var timeout time.Duration
	if err := timeoutArg(&timeout); err != nil {
		logger.Err(err.Error())
		return ExitBadArguments, err
	}

	// create db object
	db, err := sql.Open("mysql", conf.Server)
	if err != nil {
		return ExitLockFailed, err
	}
	defer db.Close()

	mutex := &MysqlMutex{
		Name: md5Hash(keyName),
		db:   db,
	}
	mutex.SetDuration(timeout)

	// Try to get the lock, block until we get the lock or we reached the timeout value
	if err = mutex.Lock(); err != nil {
		return ExitLockFailed, err
	}
	defer mutex.Unlock()

	// create a Cmd
	cmd := exec.Command(os.Args[2:][0])
	cmd.Args = os.Args[2:]

	// Ensure that the stdout and stderr from the command gets displayed
	err = pipeCommandOutput(cmd)
	if err != nil {
		return ExitCommandRunFailed, err
	}

	// run the command and return it's exit status
	exitStatus, err := runCommand(cmd)
	return exitStatus, err
}

// RunCommand starts the command and waits until it's finished
func runCommand(cmd *exec.Cmd) (existatus int, err error) {
	// generic failure
	if err := cmd.Start(); err != nil {
		return ExitCommandRunFailed, err
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus(), nil
			}
		}
		return ExitCommandRunFailed, err
	}
	return ExitSuccess, nil
}

// PipeCommandOutput ensures that the commands output gets piped to seras output
func pipeCommandOutput(cmd *exec.Cmd) (err error) {

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	return nil
}

// timeoutArg get the first argument to sera and returns it as an integer
func timeoutArg(timeout *time.Duration) (err error) {
	args := os.Args[1:]
	seconds, err := strconv.ParseInt(args[0], 10, 0)
	*timeout = time.Duration(seconds) * time.Second
	return err
}

func md5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func log(msg interface{}) {
	if msg == nil || msg == "" {
		return
	}
	if conf.Verbose {
		fmt.Printf("%s\n", msg)
	}
	cmd := strings.Join(os.Args, " ")
	if conf.Syslog {
		switch msg.(type) {
		case error:
			logger.Err(cmd + " | " + msg.(error).Error())
		case string:
			logger.Err(cmd + " | " + msg.(string))
		}
	}
}
