package redisdata
import (
	"github.com/garyburd/redigo/redis"
	"github.com/Jumpscale/agentcontroller8/redisdata/ds"
	"github.com/Jumpscale/agentcontroller8/core"
)

type loggedCommands struct {
	connPool *redis.Pool
	redisQueue ds.CommandList
}

func NewCommandLog(connPool *redis.Pool) core.CommandLog {
	return &loggedCommands{
		connPool: connPool,
		redisQueue: ds.CommandList{List: ds.List{Value: ds.Value{Name: "cmds.log.queue"}}},
	}
}

func (logger *loggedCommands) Push(command *core.Command) error {
	return logger.redisQueue.RightPush(logger.connPool, command)
}

func (logger *loggedCommands) BlockingPop() (*core.Command, error) {
	return logger.redisQueue.BlockingLeftPop(logger.connPool, 0)
}