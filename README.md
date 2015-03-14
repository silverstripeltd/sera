# sera

Distributed Mutex locking using a mysql database

## Introduction

`sera` can stop commands from running at the same time in clustered environment.
  
For example you might have two servers running two task at the same time and this can cause
race conditions. This is normally prevented via a Message Queue (MQ) system, but there are
cases where using a MQ is over

`sera` relies on the MySQL `get_lock()` function to ensure that only one instance in a cluster
is running a command at any time. `get_lock()` is typically not supported in a master-master or 
master-slave environment.

## Usage

	sera <wait-time-in-seconds> <command to run> < .. arguments and flags to command>


### wait-time-in-seconds

This is how many seconds sera will wait for a lock to be released until it gives up and aborts
running the command. This number can be 0. 

<<<<<<< HEAD
=======
sera takes two arguments, the first one is how the is how long in seconds it would take to complete the task (upper bound). After this time another instance of sera will be able to run this job.
>>>>>>> 53a5e4b87bde87e2670efc6bf8b280da643b3d37

### command to run and flags

The second and subsequent arguments is what command sera will execute. It will use the name of 
the commands and arguments as the name for which the lock.
 
## Example

![Sera example](sera.png)


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
