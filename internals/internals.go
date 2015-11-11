// Internal commands that get executed on AgentController itself instead of being dispatched to connected Agent
// instances.
package internals
import (
	"github.com/Jumpscale/agentcontroller2/messages"
	"github.com/Jumpscale/agentcontroller2/core"
	"time"
	"encoding/json"
)

type CommandName string
type CommandFunc func(*Manager, *messages.CommandMessage) (interface{}, error)

type Manager struct {
	commandProcessors map[CommandName]CommandFunc
	agents            core.AgentInformationStorage
	outgoingSignals   messages.Outgoing
	messageResponder  func(*messages.CommandResultMessage) error
}

func NewManager(agents core.AgentInformationStorage, outgoingSignals messages.Outgoing,
	messageResponder func(*messages.CommandResultMessage) error) *Manager {

	return &Manager {
		commandProcessors: map[CommandName]CommandFunc {
			"list_agents": listAgentsCommand,
		},
		agents: agents,
		outgoingSignals: outgoingSignals,
		messageResponder: messageResponder,
	}
}

func (manager *Manager) RegisterProcessor(command CommandName, processor CommandFunc) {
	manager.commandProcessors[command] = processor
}

func (manager *Manager) ProcessInternalCommand(commandMessage *messages.CommandMessage) {

	command := commandMessage.Content

	result := &core.CommandResult{
		ID:        command.ID,
		Gid:       command.Gid,
		Nid:       command.Nid,
		Tags:      command.Tags,
		State:     core.CommandStateError,
		StartTime: int64(time.Duration(time.Now().UnixNano()) / time.Millisecond),
	}

	processor, ok := manager.commandProcessors[CommandName(command.Args.Name)]
	if ok {
		data, err := processor(manager, commandMessage)
		if err != nil {
			result.Data = err.Error()
		} else {
			serialized, err := json.Marshal(data)
			if err != nil {
				result.Data = err.Error()
			}
			result.State = core.CommandStateSuccess
			result.Data = string(serialized)
			result.Level = 20
		}
	} else {
		result.State = core.CommandStateErrorUnknownCommand
	}

	resultMessage, err := messages.CommandResultMessageFromCommandResult(result)
	if err != nil {
		panic(err)
	}

	manager.messageResponder(resultMessage)
	manager.outgoingSignals.SignalAsQueued(commandMessage)
}
