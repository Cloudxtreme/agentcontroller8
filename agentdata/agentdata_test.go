package agentdata_test
import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/Jumpscale/agentcontroller8/core"
	"github.com/Jumpscale/agentcontroller8/agentdata"
)

// Note: These tests don't validate the thread-safety aspect of the implementation

func TestAgentData(t *testing.T) {

	d := agentdata.NewAgentData()

	assert.Empty(t, d.ConnectedAgents())

	id := core.AgentID{GID: 0, NID: 42}
	id2 := core.AgentID{GID: 0, NID: 23}

	assert.Nil(t, d.GetRoles(id))
	assert.Nil(t, d.GetRoles(id2))
	assert.False(t, d.IsConnected(id))
	assert.False(t, d.IsConnected(id2))

	dummyRoles := []core.AgentRole {"dummy", "slave"}

	d.SetRoles(core.AgentID{GID: 0, NID: 42}, dummyRoles)
	assert.True(t, d.IsConnected(id))

	assert.Equal(t, d.GetRoles(id), dummyRoles)
	assert.Equal(t, d.ConnectedAgents(), []core.AgentID{id})

	dummyRoles2 := []core.AgentRole {"node", "super"}

	d.SetRoles(core.AgentID{GID: 0, NID: 23}, dummyRoles2)
	assert.True(t, d.IsConnected(id))
	assert.True(t, d.IsConnected(id2))

	assert.Equal(t, d.GetRoles(id2), dummyRoles2)
	assert.Equal(t, d.GetRoles(id), dummyRoles)
	assert.Contains(t, d.ConnectedAgents(), id)
	assert.Contains(t, d.ConnectedAgents(), id2)

	assert.True(t, d.HasRole(id, "slave"))
	assert.True(t, d.HasRole(id, "dummy"))
	assert.False(t, d.HasRole(id, "super"))
	assert.True(t, d.HasRole(id2, "super"))
	assert.True(t, d.HasRole(id2, "node"))
	assert.False(t, d.HasRole(id2, "slave"))

	assert.False(t, d.HasRole(core.AgentID{GID: 1, NID: 42}, "node"))
}


func TestQueryingForConnectedAgentsWithFilters(t *testing.T) {

	d := agentdata.NewAgentData()

	id0 := core.AgentID{GID: 0, NID: 1}
	id1 := core.AgentID{GID: 0, NID: 2}
	id2 := core.AgentID{GID: 1, NID: 0}
	id3 := core.AgentID{GID: 1, NID: 1}

	d.SetRoles(id0, []core.AgentRole{"node", "cpu", "super"})
	d.SetRoles(id1, []core.AgentRole{"node", "cpu", "master"})
	d.SetRoles(id2, []core.AgentRole{"net", "super"})
	d.SetRoles(id3, []core.AgentRole{"node", "super"})

	connectedWithoutFilters := d.FilteredConnectedAgents(nil, nil)
	assert.Len(t, connectedWithoutFilters, 4)

	nodeSuperAgents := d.FilteredConnectedAgents(nil, []core.AgentRole{"node", "super"})
	assert.Len(t, nodeSuperAgents, 2)
	assert.Contains(t, nodeSuperAgents, id0)
	assert.Contains(t, nodeSuperAgents, id3)

	gid0 := uint(0)
	gid0Master := d.FilteredConnectedAgents(&gid0, []core.AgentRole{"master"})
	assert.Len(t, gid0Master, 1)
	assert.Contains(t, gid0Master, id1)

	// Filtering with AGENT_ROLE_ALL
	gid1 := uint(1)
	gid1All := d.FilteredConnectedAgents(&gid1, []core.AgentRole{core.AgentRoleAll, core.AgentRole("net")})
	assert.Len(t, gid1All, 2)

	all := d.FilteredConnectedAgents(nil, []core.AgentRole{core.AgentRoleAll, core.AgentRole("net")})
	assert.Len(t, all, 4)
}

func TestSetRolesOverwritesPreviouslySetRoles(t *testing.T) {

	d := agentdata.NewAgentData()

	id := core.AgentID{GID: 1, NID: 0}
	d.SetRoles(id, []core.AgentRole{"net", "super"})

	assert.True(t, d.HasRole(id, "net"))
	assert.True(t, d.HasRole(id, "super"))
	assert.Len(t, d.GetRoles(id), 2)

	d.SetRoles(id, []core.AgentRole{"role1", "role2", "role3"})

	assert.False(t, d.HasRole(id, "net"))
	assert.False(t, d.HasRole(id, "super"))
	assert.Len(t, d.GetRoles(id), 3)
	assert.True(t, d.HasRole(id, "role1"))
	assert.True(t, d.HasRole(id, "role2"))
	assert.True(t, d.HasRole(id, "role3"))
}