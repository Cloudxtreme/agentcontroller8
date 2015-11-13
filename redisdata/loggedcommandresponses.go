package redisdata
import (
	"github.com/garyburd/redigo/redis"
	"github.com/Jumpscale/agentcontroller2/redisdata/ds"
	"github.com/Jumpscale/agentcontroller2/core"
)

type loggedCommandResponses struct {
	connPool *redis.Pool
	redisQueue ds.CommandResultList
}

func LoggedCommandResponse(connPool *redis.Pool) core.LoggedCommandResponses {
	return &loggedCommandResponses{
		connPool: connPool,
		redisQueue: ds.CommandResultList{List: ds.List{Name: "results.queue"}},
	}
}

func (logger *loggedCommandResponses) Push(commandResult *core.CommandResponse) error {
	return logger.redisQueue.RightPush(logger.connPool, commandResult)
}

func (logger *loggedCommandResponses) Pop() (*core.CommandResponse, error) {
	return logger.redisQueue.BlockingLeftPop(logger.connPool, 0)
}
