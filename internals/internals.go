// Internal commands that get executed on AgentController itself instead of being dispatched to connected Agent
// instances.
package internals
import (
	"github.com/Jumpscale/agentcontroller8/core"
	"github.com/Jumpscale/agentcontroller8/scheduling"
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

func (manager *Manager) ExecuteInternalCommand(command *core.Command) {

	var response *core.CommandResponse = nil

	handler, ok := manager.commandHandlers[InternalCommandName(command.Content.Args.Name)]
	if ok {
		data, err := handler(command)
		if err != nil {
			response = core.ErrorResponseFor(command, err.Error())
		} else {
			response = core.SuccessResponseFor(command, data, 20)
		}
	} else {
		response = core.UnknownCommandResponseFor(command)
	}

	manager.commandResponder.RespondToCommand(response)
	manager.commandResponder.SignalAsPickedUp(command)
}
