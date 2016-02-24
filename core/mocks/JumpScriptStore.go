package mocks

import "github.com/Jumpscale/agentcontroller8/core"
import "github.com/stretchr/testify/mock"

type JumpScriptStore struct {
	mock.Mock
}

func (_m *JumpScriptStore) Add(_a0 core.JumpScriptContent) (core.JumpScriptID, error) {
	ret := _m.Called(_a0)

	var r0 core.JumpScriptID
	if rf, ok := ret.Get(0).(func(core.JumpScriptContent) core.JumpScriptID); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(core.JumpScriptID)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(core.JumpScriptContent) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *JumpScriptStore) Get(_a0 core.JumpScriptID) (core.JumpScriptContent, error) {
	ret := _m.Called(_a0)

	var r0 core.JumpScriptContent
	if rf, ok := ret.Get(0).(func(core.JumpScriptID) core.JumpScriptContent); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(core.JumpScriptContent)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(core.JumpScriptID) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
