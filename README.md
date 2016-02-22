# AgentController2 #
[![Build Status](https://travis-ci.org/Jumpscale/agentcontroller2.svg?branch=master)](https://travis-ci.org/Jumpscale/agentcontroller2)

For more information checkout the [docs](https://gig.gitbooks.io/jumpscale8/content/MultiNode/AgentController2/AgentController2.html#).

JumpScale Agentcontroller in Go.

### Functions of a Agent2
- is a process manager
- a remote command executor that gets it's jobs and tasks by polling from AC (Agent Controller 2).
- tcp portforwarder
- statistics aggregator & forwarder

### Generic features
- uses SSL with client certificates for security
- out stdour/err of subprocesses is being parsed & forwarded to agentcontroller in controlled way

The Agent will also monitor the jobs, updating the AC with `stats` and `logs`. All according to specs. 

# Running from source code #
```
go run main.go -c agentcontroller2.toml
```

# Using it #
See the [Go client](/client).

# Testing #
```bash
TEST_REDIS_PORT=6379 go test ./...
```

# Architecture

![](https://docs.google.com/drawings/d/1qsOzbv2XbwChgsLVV8qCydmH0ki9QLkaB336kt7D1Cg/pub?w=960&h=720)
