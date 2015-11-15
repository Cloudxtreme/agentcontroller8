package core

// Persistent store of Agent logs
type AgentLog interface {

	// Pushes a log entry from the specified agent.
	Push(AgentID, []byte) error
}