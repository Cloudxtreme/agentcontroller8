package redisdata
import (
	"github.com/garyburd/redigo/redis"
	"github.com/Jumpscale/agentcontroller2/messages"
	"github.com/Jumpscale/agentcontroller2/redisdata/ds"
	"fmt"
)

type outgoingSignals struct {
	connPool *redis.Pool
}

func OutgoingSignals(connPool *redis.Pool) messages.OutgoingSignals {
	return &outgoingSignals{
		connPool: connPool,
	}
}

func (outgoing *outgoingSignals) redisQueue(command *messages.CommandMessage) ds.List {
	return ds.List{Name: fmt.Sprintf("cmd.%s.queued", command.Content.ID)}
}

func (outgoing *outgoingSignals) SignalAsQueued(command *messages.CommandMessage) {
	outgoing.redisQueue(command).RightPush(outgoing.connPool, []byte("queued"))
}
