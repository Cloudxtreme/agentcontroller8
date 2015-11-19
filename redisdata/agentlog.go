package redisdata
import (
	"github.com/garyburd/redigo/redis"
	"github.com/Jumpscale/agentcontroller2/core"
	"github.com/Jumpscale/agentcontroller2/redisdata/ds"
	"fmt"
)

type agentLog struct {
	connPool *redis.Pool
}

func NewAgentLog(connPool *redis.Pool) core.AgentLog {
	return &agentLog{
		connPool: connPool,
	}
}

func (log *agentLog) Push(agentID core.AgentID, entry []byte) error {
	return ds.GetList(fmt.Sprintf("%s:%s:log", agentID.GID, agentID.NID)).RightPush(log.connPool, entry)
}