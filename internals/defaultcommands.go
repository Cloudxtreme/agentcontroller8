package internals
import (
	"fmt"
	"github.com/Jumpscale/agentcontroller2/core"
)

// Caller is expecting a map with keys "GID:NID" of each live agent and values being
// the sequence of roles the agent declares.
func listAgentsCommand(manager *Manager, cmd *core.Command) (interface{}, error) {
	output := make(map[string][]string)
	for _, agentID := range manager.agents.ConnectedAgents() {
		var roles []string
		for _, role := range manager.agents.GetRoles(agentID) {
			roles = append(roles, string(role))
		}
		output[fmt.Sprintf("%d:%d", agentID.GID, agentID.NID)] = roles
	}
	return output, nil
}