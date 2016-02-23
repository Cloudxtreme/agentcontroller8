package logged_test
import (
	"testing"
	"github.com/Jumpscale/agentcontroller8/core"
	"github.com/Jumpscale/agentcontroller8/core/mocks"
	"github.com/Jumpscale/agentcontroller8/logged"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func dummyCommand() *core.Command {
	return core.CommandFromContent(&core.CommandContent{})
}

func TestItLogsAboutPoppedCommands(t *testing.T) {

	command := dummyCommand()

	commandSource := new(mocks.CommandSource)
	commandSource.On("BlockingPop").Return(command, nil)

	commandLog := new(mocks.CommandLog)
	commandLog.On("Push", mock.Anything).Return(nil)

	loggedCommandSource := &logged.CommandSource{Log: commandLog, CommandSource: commandSource}

	poppedCommand, err := loggedCommandSource.BlockingPop()

	assert.NoError(t, err)
	assert.Equal(t, *command, *poppedCommand)

	commandLog.AssertCalled(t, "Push", command)
	commandLog.AssertNumberOfCalls(t, "Push", 1)
}

func TestImplementsCommandSource(t *testing.T) {
	assert.Implements(t, (*core.CommandSource)(nil), new(logged.CommandSource))
}
