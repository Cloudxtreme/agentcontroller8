package redisdata
import (
	"github.com/garyburd/redigo/redis"
	"github.com/Jumpscale/agentcontroller2/redisdata/ds"
	"fmt"
	"github.com/Jumpscale/agentcontroller2/core"
	"time"
)

type commandResponder struct {
	connPool *redis.Pool
}

// Constructs a core.CommandResponder implementation that responds directly to a data structures on a common
// Redis server.
func NewRedisCommandResponder(connPool *redis.Pool) core.CommandResponder {
	return &commandResponder{
		connPool: connPool,
	}
}

func listForSignal(command *core.Command) ds.List {
	return ds.GetList(fmt.Sprintf("cmd.%s.queued", command.Content.ID))
}

func hashForCommandResult(commandResult *core.CommandResponse) ds.CommandResultHash {
	return ds.CommandResultHash{Hash: ds.GetHash(fmt.Sprintf("jobresult:%s", commandResult.Content.ID))}
}

func singletonListForCommandResult(result *core.CommandResponse) ds.CommandResultList {
	name := fmt.Sprintf("cmd.%s.%d.%d", result.Content.ID, result.Content.Gid, result.Content.Nid)
	return ds.GetCommandResultList(name)
}

func (outgoing *commandResponder) SignalAsPickedUp(command *core.Command) {
	listForSignal(command).RightPush(outgoing.connPool, []byte("queued"))
}

func (outgoing *commandResponder) RespondToCommand(result *core.CommandResponse) error {

	hash := hashForCommandResult(result)

	err := hash.Set(outgoing.connPool, fmt.Sprintf("%d:%d", result.Content.Gid, result.Content.Nid), result)
	if err != nil {
		return err
	}

	err = hash.Hash.Expire(outgoing.connPool, 24 * time.Hour)
	if err != nil {
		return err
	}

	if result.Content.State != core.CommandStateQueued && result.Content.State != core.CommandStateRunning {
		singletonList := singletonListForCommandResult(result)

		singletonList.RightPush(outgoing.connPool, result)
		if err != nil {
			return err
		}

		singletonList.List.Expire(outgoing.connPool, 24 * time.Hour)
		if err != nil {
			return err
		}
	}

	return nil
}