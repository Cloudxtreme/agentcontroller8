package redisdata

import (
	"fmt"
	"github.com/Jumpscale/agentcontroller8/core"
	"github.com/Jumpscale/agentcontroller8/redisdata/ds"
	"github.com/garyburd/redigo/redis"
	"log"
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

func listForPickedUpSignal(command *core.Command) ds.List {
	return ds.GetList(fmt.Sprintf("cmd.%s.queued", command.Content.ID))
}

func hashForCommandResponses(commandResult *core.CommandResponse) ds.CommandResultHash {
	return ds.CommandResultHash{Hash: ds.GetHash(fmt.Sprintf("jobresult:%s", commandResult.Content.ID))}
}

func listForDoneSignal(result *core.CommandResponse) ds.CommandResultList {
	name := fmt.Sprintf("cmd.%s.%d.%d", result.Content.ID, result.Content.Gid, result.Content.Nid)
	return ds.GetCommandResultList(name)
}

func (outgoing *commandResponder) SignalAsPickedUp(command *core.Command) {
	listForPickedUpSignal(command).RightPush(outgoing.connPool, []byte("queued"))
}

func (outgoing *commandResponder) RespondToCommand(response *core.CommandResponse) error {

	hash := hashForCommandResponses(response)

	err := hash.Set(outgoing.connPool, fmt.Sprintf("%d:%d", response.Content.Gid, response.Content.Nid), response)
	if err != nil {
		return err
	}

	err = hash.Hash.Expire(outgoing.connPool, 24*time.Hour)
	if err != nil {
		return err
	}

	// Signal as done if appropriate
	if core.IsTerminalCommandState(response.Content.State) {
		outgoing.signalAsDone(response)
	}

	return nil
}

func (outgoing *commandResponder) signalAsDone(response *core.CommandResponse) {

	list := listForDoneSignal(response)

	err := list.RightPush(outgoing.connPool, response)
	if err != nil {
		log.Fatalf("Redis error: %v", err)
	}

	list.List.Expire(outgoing.connPool, 24*time.Hour)
	if err != nil {
		log.Fatalf("Redis error: %v", err)
	}
}
