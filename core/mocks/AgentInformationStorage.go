package mocks

import "github.com/Jumpscale/agentcontroller8/core"
import "github.com/stretchr/testify/mock"

type AgentInformationStorage struct {
	mock.Mock
}

func (_m *AgentInformationStorage) SetRoles(id core.AgentID, roles []core.AgentRole) {
	_m.Called(id, roles)
}
func (_m *AgentInformationStorage) GetRoles(id core.AgentID) []core.AgentRole {
	ret := _m.Called(id)

	var r0 []core.AgentRole
	if rf, ok := ret.Get(0).(func(core.AgentID) []core.AgentRole); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]core.AgentRole)
		}
	}

	return r0
}
func (_m *AgentInformationStorage) DropAgent(id core.AgentID) {
	_m.Called(id)
}
func (_m *AgentInformationStorage) HasRole(id core.AgentID, role core.AgentRole) bool {
	ret := _m.Called(id, role)

	var r0 bool
	if rf, ok := ret.Get(0).(func(core.AgentID, core.AgentRole) bool); ok {
		r0 = rf(id, role)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}
func (_m *AgentInformationStorage) ConnectedAgents() []core.AgentID {
	ret := _m.Called()

	var r0 []core.AgentID
	if rf, ok := ret.Get(0).(func() []core.AgentID); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]core.AgentID)
		}
	}

	return r0
}
func (_m *AgentInformationStorage) FilteredConnectedAgents(gid *uint, roles []core.AgentRole) []core.AgentID {
	ret := _m.Called(gid, roles)

	var r0 []core.AgentID
	if rf, ok := ret.Get(0).(func(*uint, []core.AgentRole) []core.AgentID); ok {
		r0 = rf(gid, roles)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]core.AgentID)
		}
	}

	return r0
}
func (_m *AgentInformationStorage) IsConnected(id core.AgentID) bool {
	ret := _m.Called(id)

	var r0 bool
	if rf, ok := ret.Get(0).(func(core.AgentID) bool); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}
