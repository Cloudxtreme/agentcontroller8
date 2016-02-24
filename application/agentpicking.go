package application
import (
	"github.com/Jumpscale/agentcontroller8/core"
	"math/rand"
)


// Returns the agents for dispatching the given command to, or an error response to be responded-with immediately.
func agentsForCommand(liveAgents core.AgentInformationStorage, command *core.Command) []core.AgentID {

	if len(command.Content.Roles) > 0 {

		// Agents with the specified GID and Roles
		matchingAgents := liveAgents.FilteredConnectedAgents(command.AttachedGID(), command.AttachedRoles())
		if len(matchingAgents) == 0 {
			return matchingAgents
		}

		if command.Content.Fanout {
			return matchingAgents
		}

		randomAgent := matchingAgents[rand.Intn(len(matchingAgents))]
		return []core.AgentID{randomAgent}
	}

	// Matching with a specific GID,NID
	agentID := core.AgentID{GID: uint(command.Content.Gid), NID: uint(command.Content.Nid)}
	if !liveAgents.IsConnected(agentID) {
		// Choose none
		return []core.AgentID{}
	}

	// Choose the chosen one
	return []core.AgentID{agentID}
}
