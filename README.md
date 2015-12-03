# sera

Distributed Mutex locking using a mysql database
[![Coverage Status](https://coveralls.io/repos/silverstripe-labs/sera/badge.svg?branch=master&service=github)](https://coveralls.io/github/silverstripe-labs/sera?branch=master)
[![Build Status](https://travis-ci.org/silverstripe-labs/sera.svg?branch=master)](https://travis-ci.org/silverstripe-labs/sera)

## Introduction

`sera` stops commands from running at the same time in clustered environment.

For example you might have two servers making an attempt to run the same task at the same moment, which can cause race
conditions. This can be prevented via a Message Queue (MQ) or centralised scheduling system, but sometimes you just want
a simpler solution.

`sera` will only lock commands with exactly the same arguments from running concurrently. For example `sera script.sh 10` and
`sera script.sh 20` will not be lock each other.

`sera` relies on the MySQL `get_lock()` function to ensure that only one instance in a cluster
is running a command at any time. Note however that `get_lock()` will not work in a Galera cluster or any 
other Master <> Master set up.

## Usage

	sera <timeout-in-seconds> <command to run> < .. arguments and flags to command>

### timeout-in-seconds

Sera will wait <timeout-in-seconds> for a lock to be released. If there is no available lock within that time sera will
abort running the <command to run>.
  
If want the command to always be executed on every node, you set this value to the maximum time and some more for the command 
to be run on every node in the cluster. For example:

long-running-task.sh takes maximum 30 seconds to run on a node and you have 4 nodes in the cluster. Then you should call sera
with a 120 seconds <timeout-in-seconds>

    sera 120 long-running-task.sh

If it doesn't matter which node runs the command, but it's important that it doesn't run at the same time, you should set the 
 <timeout-in-seconds> to 0.
 
    sera 0 there-can-be-only-one.sh

### command to run & flags

The second and subsequent arguments is what command sera will execute. Sera will make an md5 hash of the commands and arguments
and use that as the lock name across the cluster.
 
## Example

These two commands were started at roughly the same time, but only the one on the left got the lock
first and the one to the right timed out after 5 seconds.

![Sera example](https://raw.githubusercontent.com/stojg/sera/master/usage.png)


## Configuration

Sera is looking for a config file located at  `/etc/sera.json`

	{
		"server": "sera:secret@tcp(127.0.0.1:3306)/?timeout=500ms",
		"syslog": true,
		"verbose": true
	}

**server**:  A Data Source Name string for connecting to a MySQL database, as described 
here (https://github.com/go-sql-driver/mysql#dsn-data-source-name)[https://github.com/go-sql-driver/mysql#dsn-data-source-name]

**syslog**: If sera should log errors and failed locking attempts to syslog

**verbose**: if sera should print log messages to stdout

## Installation:

Add the configuration file `/etc/sera.json` and either:

 - Download a binary from the (releases)[https://github.com/stojg/sera/releases]
 - Install via `go get github.com/silverstripe-labs/sera && go install github.com/silverstripe-labs/sera`


## Caveats

Be wary of normal bash syntax limitation. For example:
 
    sera 5 task.sh; echo "hello" 
    
will always execute `echo "hello"` due to the `;`

    sera 5 task.sh && dosomething.sh 
    
will only execute `dosomething.sh` if sera got a lock and task.sh succeeded.

    sera 5 task.sh || dosomething.sh
    
will only execute `dosomething.sh` if sera couldn't get a lock or if the task.sh returned a non zero exit 
code.

    sera 5 task.sh | wc 
    
will pipe the result of `task.sh` to wc