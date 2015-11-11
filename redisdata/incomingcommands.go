package redisdata
import (
	"github.com/garyburd/redigo/redis"
	"github.com/Jumpscale/agentcontroller2/redisdata/ds"
	"github.com/Jumpscale/agentcontroller2/core"
)

type incomingCommands struct {
	connPool *redis.Pool
	redisQueue	ds.CommandList
}

func IncomingCommands(connPool *redis.Pool) core.IncomingCommands {
	return &incomingCommands{
		connPool: connPool,
		redisQueue: ds.CommandList{List: ds.List{Name: "cmds.queue"}},
	}
}

func (incoming *incomingCommands) Pop() (*core.Command, error) {
	return incoming.redisQueue.BlockingLeftPop(incoming.connPool, 0)
}

func (incoming *incomingCommands) Push(command *core.Command) error {
	return incoming.redisQueue.RightPush(incoming.connPool, command)
}