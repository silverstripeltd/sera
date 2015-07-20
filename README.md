# sera

Distributed Mutex locking using a mysql database
[![Coverage Status](https://coveralls.io/repos/silverstripe-labs/sera/badge.svg?branch=master&service=github)](https://coveralls.io/github/silverstripe-labs/sera?branch=master)
[![Build Status](https://travis-ci.org/silverstripe-labs/sera.svg?branch=master)](https://travis-ci.org/silverstripe-labs/sera)

## Introduction

`sera` stops commands from running at the same time in clustered environment.

For example you might have two servers making an attempt to run the same task at the same moment, which can cause race
conditions. This is normally prevented via a Message Queue (MQ) system, but there are cases where using a MQ adds too
much overhead.

`sera` will only lock out commands with the same footprint from running concurrently. For example `script.sh 10` and
`script.sh 20` will not be protected from each other - they will run as normal.

`sera` relies on the MySQL `get_lock()` function to ensure that only one instance in a cluster
is running a command at any time. Note however that `get_lock()` will not work in a Galera cluster.

## Usage

	sera <wait-time-in-seconds> <command to run> < .. arguments and flags to command>

### wait-time-in-seconds

This is how many seconds sera will wait for a lock to be released until it gives up and aborts
running the command. This number can be 0. 

### command to run and flags

The second and subsequent arguments is what command sera will execute. It will use the name of 
the commands and arguments as the name for the lock.
 
## Example

These two commands were started at roughly the same time, but only the one on the left got the lock
first and the one to the right timed out after 5 seconds.

![Sera example](https://raw.githubusercontent.com/stojg/sera/master/usage.png)


## Configuration

`/etc/sera.json`


	{
		"server": "sera:secret@tcp(127.0.0.1:3306)/?timeout=500ms",
		"syslog": true,
		"verbose": true
	}

**server**:  A Data Source Name string for connecting to a MySQL database, as described 
here (https://github.com/go-sql-driver/mysql#dsn-data-source-name)[https://github.com/go-sql-driver/mysql#dsn-data-source-name]

**syslog**: If sera should log errors and failed locking attempts to syslog

**verbose**: if sera should print errors to stdout


## Installation:

Add the configuration file and either:

 - Download a binary from the (releases)[https://github.com/stojg/sera/releases]
 - Install via `go get github.com/stojg/sera && go install github.com/stojg/sera`
