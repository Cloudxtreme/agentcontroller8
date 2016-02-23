package application
import (
	"testing"
	"github.com/Jumpscale/agentcontroller8/core"
	"github.com/Jumpscale/agentcontroller8/agentdata"
	"github.com/stretchr/testify/assert"
)

func TestAgentPickingForCommandRouting(t *testing.T) {

	agentInformation := agentdata.NewAgentData()

	agent1 := core.AgentID{GID: 1, NID: 0}
	agentInformation.SetRoles(agent1, []core.AgentRole{"super", "node"})

	agent2 := core.AgentID{GID: 1, NID: 1}
	agentInformation.SetRoles(agent2, []core.AgentRole{"cpu", "node"})

	agent3 := core.AgentID{GID: 3, NID: 1}
	agentInformation.SetRoles(agent3, []core.AgentRole{"worker", "cpu"})

	agent4 := core.AgentID{GID: 3, NID: 2}
	agentInformation.SetRoles(agent4, []core.AgentRole{"super", "net"})

	// Matching with roles (fanning out)
	match1 := agentsForCommand(agentInformation,
		core.CommandFromContent(&core.CommandContent{Roles: []string{"super"}, Fanout: true}))
	assert.Contains(t, match1, agent1)
	assert.Contains(t, match1, agent4)

	// Matching with roles (without fanning out)
	match2 := agentsForCommand(agentInformation,
		core.CommandFromContent(&core.CommandContent{Roles: []string{"super"}}))
	assert.Len(t, match2, 1)

	// Matching a specific Agent
	match3 := agentsForCommand(agentInformation, core.CommandFromContent(&core.CommandContent{Gid: 3, Nid: 1}))
	assert.Len(t, match3, 1)
	assert.Contains(t, match3, agent3)
}
