package redisdata
import (
	"github.com/garyburd/redigo/redis"
	"github.com/Jumpscale/agentcontroller2/redisdata/ds"
	"fmt"
	"github.com/Jumpscale/agentcontroller2/core"
)

type outgoing struct {
	connPool *redis.Pool
}

func Outgoing(connPool *redis.Pool) core.Outgoing {
	return &outgoing{
		connPool: connPool,
	}
}

func listForSignal(command *core.Command) ds.List {
	return ds.List{Name: fmt.Sprintf("cmd.%s.queued", command.Content.ID)}
}

func hashForCommandResult(commandResult *core.CommandResult) ds.CommandResultHash {
	return ds.CommandResultHash{Hash: ds.Hash{Name: fmt.Sprintf("jobresult:%s", commandResult.Content.ID)}}
}

func singletonListForCommandResult(result *core.CommandResult) ds.CommandResultList {
	name := fmt.Sprintf("cmd.%s.%d.%d", result.Content.ID, result.Content.Gid, result.Content.Nid)
	return ds.CommandResultList{List: ds.List{Name: name}}
}

func (outgoing *outgoing) SignalAsQueued(command *core.Command) {
	listForSignal(command).RightPush(outgoing.connPool, []byte("queued"))
}

func (outgoing *outgoing) RespondToCommand(result *core.CommandResult) error {

	err := hashForCommandResult(result).Set(outgoing.connPool,
		fmt.Sprintf("%d:%d", result.Content.Gid, result.Content.Nid),
		result)
	if err != nil {
		return err
	}

	if result.Content.State != core.CommandStateQueued && result.Content.State != core.CommandStateRunning {
		singletonListForCommandResult(result).RightPush(outgoing.connPool, result)
		if err != nil {
			return err
		}
	}

	return nil
}