# sera

Distributed Mutex locking using a mysql database
[![Coverage Status](https://coveralls.io/repos/silverstripe-labs/sera/badge.svg?branch=master&service=github)](https://coveralls.io/github/silverstripe-labs/sera?branch=master)
[![Build Status](https://travis-ci.org/silverstripe-labs/sera.svg?branch=master)](https://travis-ci.org/silverstripe-labs/sera)

## Introduction

`sera` stops commands from running at the same time in clustered environment.

We want to prevent two servers from running the same task at the same time when
they can cause cause race conditions. This can be prevented via a Message Queue
(MQ) or centralised scheduling system.

sera is a simpler solution to preventing a race condition in a multiple server 
environment.

`sera` will only lock commands with exactly the same arguments from running 
concurrently. E.g. Running `sera script.sh 10` on server A and 
`sera script.sh 20` on server B will not be prevented from running at the same 
time.

`sera` relies on the MySQL `get_lock()` function to ensure that only one 
server in a cluster is running a command at any time. Note however that 
`get_lock()` will not work in a Galera cluster or any other Master/Master set 
up.

## Usage

	sera <timeout-in-seconds> <command to run> < .. arguments and flags to command>

### timeout-in-seconds

Sera will wait <timeout-in-seconds> for a lock to be released. If the task is
currently locked and is not release within that time period, sera will not 
execute the <command to run>.
  
If we want the command to always be executed on every node, but staggered, we 
set this value to the maximum time (and some more) for the command to be run on 
every server in the cluster. For example:

long-running-task.sh takes maximum 30 seconds to finish on a server and we have 
four servers in the cluster. We then set the timeout to 120 seconds (4 * 30sec).

    sera 120 long-running-task.sh

If we don't care which node runs the command, but it's important that it doesn't
overlap with another execution, we set the <timeout-in-seconds> to 0.
 
    sera 0 there-can-be-only-one.sh

### command to run & flags

The second and subsequent arguments is what command sera will execute. Sera will
make an md5 hash of the commands and arguments and use that as the lock name 
across the cluster.
 
## Example - Screenshot

These two commands were started at roughly the same time, but only the one on 
the left got the lock first and the one to the right timed out after 5 seconds.

![Sera example](https://raw.githubusercontent.com/stojg/sera/master/usage.png)


## Configuration

Sera needs a JSON config file located at  `/etc/sera.json`

Example: 

	{
		"server": "sera:secret@tcp(127.0.0.1:3306)/?timeout=500ms",
		"syslog": true,
		"verbose": true
	}

**server**: A Data Source Name string for connecting to a MySQL database, as 
described per (https://github.com/go-sql-driver/mysql#dsn-data-source-name)[https://github.com/go-sql-driver/mysql#dsn-data-source-name]

**syslog**: If sera should log errors and failed locking attempts to syslog

**verbose**: if sera should print log messages to stdout

## Installation:

Add the configuration file `/etc/sera.json` and either:

 - Download a binary from the (releases)[https://github.com/stojg/sera/releases]
 - Install via `go get github.com/silverstripe-labs/sera && go install github.com/silverstripe-labs/sera`


## Bash limitations and caveats 

Note normal bash syntax limitation. For example:
 
    sera 5 task.sh; echo "hello" 
    
will always execute `echo "hello"` because of the `;`

    sera 5 task.sh && dosomething.sh 
    
will only execute `dosomething.sh` if sera got a lock and task.sh succeeded.

    sera 5 task.sh || dosomething.sh
    
will only execute `dosomething.sh` if sera couldn't get a lock or if the task.sh
returned a non zero exit code.

    sera 5 task.sh | wc 
    
will pipe the stdout of `task.sh` to `wc`