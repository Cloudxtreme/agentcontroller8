package messages
import (
	"github.com/Jumpscale/agentcontroller2/redisdata/ds"
	"github.com/garyburd/redigo/redis"
)

type RedisCommandResultHash struct {
	Hash ds.Hash
}

func (hash RedisCommandResultHash) Set(connPool *redis.Pool, key string, message *CommandResultMessage) error {
	return hash.Hash.Set(connPool, key, message.Payload)
}


