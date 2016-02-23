package mocks

import "github.com/Jumpscale/agentcontroller8/core"
import "github.com/stretchr/testify/mock"

type CommandResponseLog struct {
	mock.Mock
}

func (_m *CommandResponseLog) Push(_a0 *core.CommandResponse) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*core.CommandResponse) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *CommandResponseLog) BlockingPop() (*core.CommandResponse, error) {
	ret := _m.Called()

	var r0 *core.CommandResponse
	if rf, ok := ret.Get(0).(func() *core.CommandResponse); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*core.CommandResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
