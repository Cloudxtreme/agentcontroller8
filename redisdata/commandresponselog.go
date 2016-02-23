package redisdata
import (
	"github.com/garyburd/redigo/redis"
	"github.com/Jumpscale/agentcontroller8/redisdata/ds"
	"github.com/Jumpscale/agentcontroller8/core"
)

type commandResponseLog struct {
	connPool *redis.Pool
	redisQueue ds.CommandResultList
}

func NewCommandResponseLog(connPool *redis.Pool) core.CommandResponseLog {
	return &commandResponseLog{
		connPool: connPool,
		redisQueue: ds.GetCommandResultList("results.queue"),
	}
}

func (logger *commandResponseLog) Push(commandResult *core.CommandResponse) error {
	return logger.redisQueue.RightPush(logger.connPool, commandResult)
}

func (logger *commandResponseLog) BlockingPop() (*core.CommandResponse, error) {
	return logger.redisQueue.BlockingLeftPop(logger.connPool, 0)
}
