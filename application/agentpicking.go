package application
import (
	"github.com/Jumpscale/agentcontroller2/core"
	"fmt"
	"math/rand"
)


// Returns the agents for dispatching the given command to, or an error response to be responded-with immediately.
func agentsForCommand(liveAgents core.AgentInformationStorage,
	command *core.Command) ([]core.AgentID, *core.CommandResult) {

	if len(command.Content.Roles) > 0 {

		// Agents with the specified GID and Roles
		matchingAgents := liveAgents.FilteredConnectedAgents(command.AttachedGID(), command.AttachedRoles())

		if len(matchingAgents) == 0 {
			// None chosen.
			// Respond with error immediately
			errorResponse := errorResponseFor(command, fmt.Sprintf("No agents with role '%v' alive!", command.Content.Roles))
			return []core.AgentID{}, errorResponse
		} else {
			if command.Content.Fanout {
				return matchingAgents, nil
			} else {
				randomAgent := matchingAgents[rand.Intn(len(matchingAgents))]
				return []core.AgentID{randomAgent}, nil
			}
		}
	} else {
		// Matching with a specific GID,NID
		agentID := core.AgentID{GID: uint(command.Content.Gid), NID: uint(command.Content.Nid)}
		if !liveAgents.IsConnected(agentID) {
			// Respond with error
			// Choose none
			errorResponse := errorResponseFor(command, fmt.Sprintf("Agent is not alive!"))
			return []core.AgentID{}, errorResponse
		} else {
			// Choose the chosen one
			return []core.AgentID{agentID}, nil
		}
	}
}