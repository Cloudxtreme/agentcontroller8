package mocks

import "github.com/Jumpscale/agentcontroller8/core"
import "github.com/stretchr/testify/mock"

type CommandResponder struct {
	mock.Mock
}

func (_m *CommandResponder) SignalAsPickedUp(_a0 *core.Command) {
	_m.Called(_a0)
}
func (_m *CommandResponder) RespondToCommand(_a0 *core.CommandResponse) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*core.CommandResponse) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
