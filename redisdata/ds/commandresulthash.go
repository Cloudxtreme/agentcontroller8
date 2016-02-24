package ds
import (
	"github.com/garyburd/redigo/redis"
	"github.com/Jumpscale/agentcontroller8/core"
)

type CommandResultHash struct {
	Hash Hash
}

func GetCommandResultHash(name string) CommandResultHash {
	return CommandResultHash{GetHash(name)}
}

func (hash CommandResultHash) Set(connPool *redis.Pool, key string, message *core.CommandResponse) error {
	return hash.Hash.Set(connPool, key, message.JSON)
}


