// Internal commands that get executed on AgentController itself instead of being dispatched to connected Agent
// instances.
package internals
import (
	"github.com/Jumpscale/agentcontroller2/core"
	"time"
	"encoding/json"
	"github.com/Jumpscale/agentcontroller2/scheduling"
)

type InternalCommandName string
type CommandHandler func(*core.Command) (interface{}, error)

const (
	ListAgents = InternalCommandName("list_agents")
	SchedulerAddJob = InternalCommandName("scheduler_add")
	SchedulerListJobs = InternalCommandName("scheduler_list")
	SchedulerRemoveJob = InternalCommandName("scheduler_remove")
	SchedulerRemoveJobByIdPrefix = InternalCommandName("scheduler_remove_prefix")

)

type Manager struct {
	commandHandlers  map[InternalCommandName]CommandHandler
	commandResponder core.CommandResponder
}

func NewManager(agents core.AgentInformationStorage,
	scheduler *scheduling.Scheduler,
	commandResponder core.CommandResponder) *Manager {

	manager := &Manager{
		commandHandlers: map[InternalCommandName]CommandHandler{},
		commandResponder: commandResponder,
	}

	manager.setUpAgentCommands(agents)
	manager.setUpSchedulerCommands(scheduler)

	return manager
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

	processor, ok := manager.commandHandlers[InternalCommandName(command.Args.Name)]
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
