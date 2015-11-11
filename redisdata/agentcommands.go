package redisdata
import (
	"github.com/garyburd/redigo/redis"
	"github.com/Jumpscale/agentcontroller2/messages"
	"github.com/Jumpscale/agentcontroller2/core"
	"fmt"
	"github.com/Jumpscale/agentcontroller2/redisdata/ds"
)

type agentCommands struct {
	pool *redis.Pool
}

func AgentCommands(connPool *redis.Pool) messages.AgentCommands {
	return &agentCommands{
		pool: connPool,
	}
}

func (commands *agentCommands) Enqueue(agentID core.AgentID, command *messages.CommandMessage) error {
	return commands.redisQueue(agentID).RightPush(commands.pool, command)
}

func (commands *agentCommands) Dequeue(agentID core.AgentID) (*messages.CommandMessage, error) {
	return commands.redisQueue(agentID).BlockingLeftPop(commands.pool, 0)
}

func (commands *agentCommands) ReportUnexecutedCommand(command *messages.CommandMessage, agentID core.AgentID) error {
	return commands.redisQueue(agentID).RightPush(commands.pool, command)
}

func (commands *agentCommands) redisQueue(id core.AgentID) ds.CommandList {
	name := fmt.Sprintf("cmds:%d:%d", id.GID, id.NID)
	return ds.CommandList{List: ds.List{Name: name}}
}