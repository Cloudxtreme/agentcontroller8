package logged_test
import (
	"testing"
	"github.com/Jumpscale/agentcontroller8/core"
	"github.com/Jumpscale/agentcontroller8/core/mocks"
	"github.com/Jumpscale/agentcontroller8/logged"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func dummyCommandResponse() *core.CommandResponse {
	return core.CommandResponseFromContent(&core.CommandResponseContent{})
}

func TestItLogsAboutResponses(t *testing.T) {

	response := dummyCommandResponse()

	responder := new(mocks.CommandResponder)
	responder.On("RespondToCommand", mock.Anything).Return(nil)

	responseLog := new(mocks.CommandResponseLog)
	responseLog.On("Push", mock.Anything).Return(nil)

	loggedCommandResponder := &logged.CommandResponder{Log: responseLog, CommandResponder: responder}

	err := loggedCommandResponder.RespondToCommand(response)

	assert.NoError(t, err)

	responseLog.AssertCalled(t, "Push", response)
	responseLog.AssertNumberOfCalls(t, "Push", 1)

	responder.AssertCalled(t, "RespondToCommand", response)
	responder.AssertNumberOfCalls(t, "RespondToCommand", 1)
}

func TestImplementsCommandResponder(t *testing.T) {
	assert.Implements(t, (*core.CommandResponder)(nil), new(logged.CommandResponder))
}