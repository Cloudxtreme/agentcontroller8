package internals
import (
	"fmt"
	"github.com/Jumpscale/agentcontroller8/core"
)


func (manager *Manager) setUpAgentCommands(agentInfo core.AgentInformationStorage) {

	// Caller is expecting a map with keys "GID:NID" of each live agent and values being
	// the sequence of roles the agent declares.
	manager.commandHandlers[ListAgents] = func(_ *core.Command) (interface{}, error) {
		output := make(map[string][]string)
		for _, agentID := range agentInfo.ConnectedAgents() {
			var roles []string
			for _, role := range agentInfo.GetRoles(agentID) {
				roles = append(roles, string(role))
			}
			output[fmt.Sprintf("%d:%d", agentID.GID, agentID.NID)] = roles
		}
		return output, nil
	}
}