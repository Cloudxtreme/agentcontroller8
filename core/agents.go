package core

type AgentID struct {
	GID uint	`json:"gid"`
	NID uint	`json:"nid"`
}

type AgentRole string

const AgentRoleAll = AgentRole("*")

// Information about connected Agents
type AgentInformationStorage interface {

	// Sets the roles associated with an Agent, overwriting any previously-set roles.
	SetRoles(id AgentID, roles []AgentRole)

	// Gets the roles associated with an Agent
	GetRoles(id AgentID) []AgentRole

	// Drops all the known information about an Agent
	DropAgent(id AgentID)

	// Checks if the specified agent has the specified role
	HasRole(id AgentID, role AgentRole) bool

	// Queries for all the available Agents
	ConnectedAgents() []AgentID

	// Queries for all the available agents that specify the given criteria:
	//	- If gid is not nil, only returns IDs of Agents with that GID
	//	- if roles is not nil, only returns IDs of Agents that have all of these roles
	//	- if roles include AGENT_ROLE_ALL do not filter by roles at all
	FilteredConnectedAgents(gid *uint, roles []AgentRole) []AgentID

	IsConnected(id AgentID) bool
}