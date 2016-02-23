package mocks

import "github.com/Jumpscale/agentcontroller8/core"
import "github.com/stretchr/testify/mock"

type CommandLog struct {
	mock.Mock
}

func (_m *CommandLog) Push(_a0 *core.Command) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*core.Command) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *CommandLog) BlockingPop() (*core.Command, error) {
	ret := _m.Called()

	var r0 *core.Command
	if rf, ok := ret.Get(0).(func() *core.Command); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*core.Command)
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
