package redisdata
import (
	"github.com/Jumpscale/agentcontroller2/messages"
	"github.com/garyburd/redigo/redis"
	"github.com/Jumpscale/agentcontroller2/redisdata/ds"
)

type loggedCommands struct {
	connPool *redis.Pool
	redisQueue messages.RedisCommandList
}

func LoggedCommands(connPool *redis.Pool) messages.LoggedCommands {
	return &loggedCommands{
		connPool: connPool,
		redisQueue: messages.RedisCommandList{ds.List{Name: "cmds.log.queue"}},
	}
}

func (logger *loggedCommands) Push(message *messages.CommandMessage) error {
	return logger.redisQueue.RightPush(logger.connPool, message)
}

func (logger *loggedCommands) Pop() (*messages.CommandMessage, error) {
	return logger.redisQueue.BlockingLeftPop(logger.connPool, 0)
}