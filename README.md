# agentcontroller2
[![Build Status](https://travis-ci.org/Jumpscale/agentcontroller2.svg?branch=master)](https://travis-ci.org/Jumpscale/agentcontroller2)

JumpScale agentcontroller2 in Go

# Installation
```
go get github.com/Jumpscale/agentcontroller2
```

# Running jsagencontroller
```
go run main.go -c agentcontroller2.toml
```

# Testing #
```bash
TEST_REDIS_PORT=6379 go test ./...
```
