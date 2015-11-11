package messages
import "github.com/Jumpscale/agentcontroller2/core"

// Temporarily-stored commands for all agents
type AgentCommands interface {

	// Enqueues a command for an Agent's execution queue
	Enqueue(agentID core.AgentID, command *CommandMessage) error

	// Dequeues a command from an Agent's execution queue
	Dequeue(agentID core.AgentID) (*CommandMessage, error)

	// Reports a command that was dequeued for an Agent but was failed to be executed for
	// some reason or another
	ReportUnexecutedCommand(command *CommandMessage, agentID core.AgentID) error
}
