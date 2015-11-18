package application
import (
	"github.com/Jumpscale/agentcontroller2/core"
	"math/rand"
)


// Returns the agents for dispatching the given command to, or an error response to be responded-with immediately.
func agentsForCommand(liveAgents core.AgentInformationStorage, command *core.Command) []core.AgentID {

	if len(command.Content.Roles) > 0 {

		// Agents with the specified GID and Roles
		matchingAgents := liveAgents.FilteredConnectedAgents(command.AttachedGID(), command.AttachedRoles())

		if command.Content.Fanout {
			return matchingAgents
		} else {
			randomAgent := matchingAgents[rand.Intn(len(matchingAgents))]
			return []core.AgentID{randomAgent}
		}

	} else {
		// Matching with a specific GID,NID
		agentID := core.AgentID{GID: uint(command.Content.Gid), NID: uint(command.Content.Nid)}
		if !liveAgents.IsConnected(agentID) {
			// Choose none
			return []core.AgentID{}
		} else {
			// Choose the chosen one
			return []core.AgentID{agentID}
		}
	}
}