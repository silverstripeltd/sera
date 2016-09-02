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

	sera <wait-time-in-seconds> <command>
	sera --wait-and-skip -- <wait-time-in-seconds> <command>

### "wait-time-in-seconds"

Sera will wait `<timeout-in-seconds>` for a lock to be released. If the task is
currently locked and is not released within that time period, sera will not
execute the `<command>` and will return an error code.

### "command"

The second and subsequent arguments is what command sera will execute. Sera will
make an md5 hash of the commands and arguments and use that as the lock name 
across the cluster.

## Use-cases

### Staggered execution

If you need the command to execute on all nodes in a staggered manner, set the `<wait-time-in-seconds>` to the maximum
time the command will take to execute on N-1 nodes (Nth node will need to wait until all previous nodes succeed before
it can start).

So for example if a task takes 30 seconds at most, and the cluster has 4 nodes, set the timeout to at least
(4-1)x30=90.

	sera 90 autoexec.bat

If timeout is reached, the exit code will be set to 204. This exit code should not be ignored, because it signifies
the node was not able to run the command at all.

### Execute once

If the command needs to run at least once, but we don't care on which node and we are happy to proceed immediately,
set the `<wait-time-in-seconds>` to 0.

    sera 0 DOS=HIGH,UMB

This will cause sera to exit immediately if the command is already running elsewhere. The exit code will be 204, which
can be safely discarded in this use-case.

Please see caveat below for assurances made on how many times the command will actually execute.

### Execute once, others wait and skip

If we want the command to execute once, and yet we want to block all nodes until the execution completes, it's
appropriate to use the `--wait-and-skip` flag. Only the first node will actually perform the operation. All
other nodes will wait for it to finish, then report success immediately (the lock will still be acquired, but
then released immediately).

The timeout should be set to the maximum time the command will take to execute on a single node (other nodes will take
~zero time to execute). For example if the command takes 30 seconds at most, set the `<wait-time-in-seconds>` to 30
regardless of the amount of nodes in the cluster.

    sera --wait-and-skip -- 30 SET BLASTER=A220 I5 D1 H6 T6 P330 E620

If timeout is reached, the exit code will be set to 204. This code should not be ignored, because we cannot ensure at
this stage that the command has succeeded on the first node past the post.

Please see caveat below for assurances made on how many times the command will actually execute.

### Caveat: assurances on execution count

In "execute once" and "execute once, others wait and skip" use-cases, we cannot guarantee the command will actually only
execute once.  This is because we cannot know if some other node has already finished processing (and released the
lock). We can only tell if the execution is currently on-going. In other words sera works best if all nodes start
processing more or less at the same time. If one node is significantly late (or early) to the critical section, it will
cause a stray execution.

This is generally fine if the command is re-entrant (i.e. no harm running it twice), and the longer the command takes
to complete, to smaller the chance a stray execution will occur. If you you need to guarantee a single execution, you
will need to use other synchronisation methods.

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
described per [https://github.com/go-sql-driver/mysql#dsn-data-source-name](https://github.com/go-sql-driver/mysql#dsn-data-source-name)

**syslog**: If sera should log errors and failed locking attempts to syslog

**verbose**: if sera should print log messages to stdout

## Installation:

Add the configuration file `/etc/sera.json` and either:

 - Download a binary from the [releases](https://github.com/silverstripe-labs/sera/releases)
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
