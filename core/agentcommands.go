package core

// Temporarily-stored commands for all agents
type AgentCommands interface {

	// Enqueues a command for an Agent's execution queue
	Enqueue(agentID AgentID, command *Command) error

	// Dequeues a command from an Agent's execution queue
	BlockingDequeue(agentID AgentID) (*Command, error)

	// Reports a command that was dequeued for an Agent but was failed to be executed for
	// some reason or another
	ReportUnexecutedCommand(command *Command, agentID AgentID) error
}
