package redisdata
import (
	"github.com/garyburd/redigo/redis"
	"github.com/Jumpscale/agentcontroller2/messages"
	"github.com/Jumpscale/agentcontroller2/redisdata/ds"
)

type loggedCommandResults struct {
	connPool *redis.Pool
	redisQueue ds.CommandResultList
}

func LoggedCommandResult(connPool *redis.Pool) messages.LoggedCommandResults {
	return &loggedCommandResults{
		connPool: connPool,
		redisQueue: ds.CommandResultList{List: ds.List{Name: "results.queue"}},
	}
}

func (logger *loggedCommandResults) Push(commandResult *messages.CommandResultMessage) error {
	return logger.redisQueue.RightPush(logger.connPool, commandResult)
}

func (logger *loggedCommandResults) Pop() (*messages.CommandResultMessage, error) {
	return logger.redisQueue.BlockingLeftPop(logger.connPool, 0)
}
