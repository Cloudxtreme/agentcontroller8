package redisdata
import (
	"github.com/garyburd/redigo/redis"
	"github.com/Jumpscale/agentcontroller2/redisdata/ds"
	"github.com/Jumpscale/agentcontroller2/core"
)

type commandSource struct {
	connPool *redis.Pool
	redisQueue	ds.CommandList
}

// Constructs a core.CommandSource implementation that only pops commands off of a queue from a shared
// Redis server.
func NewCommandSource(connPool *redis.Pool) core.CommandSource {
	return &commandSource{
		connPool: connPool,
		redisQueue: ds.CommandList{List: ds.List{Name: "cmds.queue"}},
	}
}

func (incoming *commandSource) Pop() (*core.Command, error) {
	return incoming.redisQueue.BlockingLeftPop(incoming.connPool, 0)
}

func (incoming *commandSource) Push(command *core.Command) error {
	return incoming.redisQueue.RightPush(incoming.connPool, command)
}