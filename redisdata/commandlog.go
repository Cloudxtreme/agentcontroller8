package redisdata
import (
	"github.com/garyburd/redigo/redis"
	"github.com/Jumpscale/agentcontroller2/redisdata/ds"
	"github.com/Jumpscale/agentcontroller2/core"
)

type loggedCommands struct {
	connPool *redis.Pool
	redisQueue ds.CommandList
}

func NewCommandLog(connPool *redis.Pool) core.CommandLog {
	return &loggedCommands{
		connPool: connPool,
		redisQueue: ds.CommandList{ds.List{Name: "cmds.log.queue"}},
	}
}

func (logger *loggedCommands) Push(command *core.Command) error {
	return logger.redisQueue.RightPush(logger.connPool, command)
}

func (logger *loggedCommands) Pop() (*core.Command, error) {
	return logger.redisQueue.BlockingLeftPop(logger.connPool, 0)
}