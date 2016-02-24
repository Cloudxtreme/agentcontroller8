package mocks

import "github.com/Jumpscale/agentcontroller8/core"
import "github.com/stretchr/testify/mock"

type AgentCommands struct {
	mock.Mock
}

func (_m *AgentCommands) Enqueue(agentID core.AgentID, command *core.Command) error {
	ret := _m.Called(agentID, command)

	var r0 error
	if rf, ok := ret.Get(0).(func(core.AgentID, *core.Command) error); ok {
		r0 = rf(agentID, command)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *AgentCommands) BlockingDequeue(agentID core.AgentID) (*core.Command, error) {
	ret := _m.Called(agentID)

	var r0 *core.Command
	if rf, ok := ret.Get(0).(func(core.AgentID) *core.Command); ok {
		r0 = rf(agentID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*core.Command)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(core.AgentID) error); ok {
		r1 = rf(agentID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *AgentCommands) ReportUnexecutedCommand(command *core.Command, agentID core.AgentID) error {
	ret := _m.Called(command, agentID)

	var r0 error
	if rf, ok := ret.Get(0).(func(*core.Command, core.AgentID) error); ok {
		r0 = rf(command, agentID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
