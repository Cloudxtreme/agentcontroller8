package redisdata
import (
	"github.com/garyburd/redigo/redis"
	"github.com/Jumpscale/agentcontroller2/messages"
	"github.com/Jumpscale/agentcontroller2/redisdata/ds"
)

type incomingCommands struct {
	connPool *redis.Pool
	redisQueue	messages.RedisCommandList
}

func IncomingCommands(connPool *redis.Pool) messages.IncomingCommands {
	return &incomingCommands{
		connPool: connPool,
		redisQueue: messages.RedisCommandList{List: ds.List{Name: "cmds.queue"}},
	}
}

func (incoming *incomingCommands) Pop() (*messages.CommandMessage, error) {
	return incoming.redisQueue.BlockingPop(incoming.connPool, 0)
}

func (incoming *incomingCommands) Push(command *messages.CommandMessage) error {
	return incoming.redisQueue.RightPush(incoming.connPool, command)
}