package ds
import (
	"github.com/garyburd/redigo/redis"
	"github.com/Jumpscale/agentcontroller2/core"
)

type CommandResultHash struct {
	Hash Hash
}

func (hash CommandResultHash) Set(connPool *redis.Pool, key string, message *core.CommandResult) error {
	return hash.Hash.Set(connPool, key, message.JSON)
}


