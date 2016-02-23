package mocks

import "github.com/Jumpscale/agentcontroller8/core"
import "github.com/stretchr/testify/mock"

type AgentLog struct {
	mock.Mock
}

func (_m *AgentLog) Push(_a0 core.AgentID, _a1 []byte) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(core.AgentID, []byte) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
