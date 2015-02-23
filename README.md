# sera

Distributed Mutex locking using a redis.

## Introduction

Sera allows mutual exclusive locking for a cluster of servers. It's normal use case is to
prevent several cronjobs or scheduled tasks running at the same time.

> Distributed locks are a very useful primitive in many environments where different 
> processes require to operate with shared resources in a mutually exclusive way.
>
> [Distributed locks with Redis](http://redis.io/topics/distlock)

## Usage

The normal use case is in a cronjob or scheduled services

	sera <expiry in seconds> <command to run> < .. arguments and flags to command>

Example cronjob:

	* * * * * root /usr/local/bin/sera 20 /bin/long-running-task --parameter hello

sera takes two arguments, the first one is how many seconds the task should take roughly. This 
will internally translate into an expiry time for this task.

If the second argument (which is the command to run) takes longer than this time, other servers 
might expire the lock and start the task.


## Resources

[http://redis.io/topics/distlock](http://redis.io/topics/distlock)

## Todo

 - Configurable redis endpoint
 - Configurable # of retries
 - Configurable delay between retries
 - Syslog logging
 - Warnings if all redis servers are unreachable 


