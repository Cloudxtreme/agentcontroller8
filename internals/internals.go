// Internal commands that get executed on AgentController itself instead of being dispatched to connected Agent
// instances.
package internals
import (
	"github.com/Jumpscale/agentcontroller2/core"
	"time"
	"encoding/json"
	"github.com/Jumpscale/agentcontroller2/scheduling"
)

type CommandHandler func(*core.Command) (interface{}, error)

type Manager struct {
	commandHandlers  map[core.CommandName]CommandHandler
	commandResponder core.CommandResponder
}

func NewManager(agents core.AgentInformationStorage,
	scheduler *scheduling.Scheduler,
	commandResponder core.CommandResponder) *Manager {

	manager := &Manager{
		commandHandlers: map[core.CommandName]CommandHandler{},
		commandResponder: commandResponder,
	}

	manager.setUpAgentCommands(agents)
	manager.setUpSchedulerCommands(scheduler)

	return manager
}

func (manager *Manager) registerProcessor(command core.CommandName, processor CommandHandler) {
	manager.commandHandlers[command] = processor
}

func (manager *Manager) ExecuteInternalCommand(commandMessage *core.Command) {

	command := commandMessage.Content

	result := &core.CommandResponseContent{
		ID:        command.ID,
		Gid:       command.Gid,
		Nid:       command.Nid,
		Tags:      command.Tags,
		State:     core.CommandStateError,
		StartTime: int64(time.Duration(time.Now().UnixNano()) / time.Millisecond),
	}

	processor, ok := manager.commandHandlers[core.CommandName(command.Args.Name)]
	if ok {
		data, err := processor(commandMessage)
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

	resultMessage := core.CommandResponseFromContent(result)

	manager.commandResponder.RespondToCommand(resultMessage)
	manager.commandResponder.SignalAsPickedUp(commandMessage)
}
