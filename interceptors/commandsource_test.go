package interceptors_test
import (
	"testing"
	"github.com/Jumpscale/agentcontroller8/inmemorydata"
	"github.com/Jumpscale/agentcontroller8/core/mocks"
	"github.com/Jumpscale/agentcontroller8/core"
	"github.com/Jumpscale/agentcontroller8/utils"
	"github.com/Jumpscale/agentcontroller8/interceptors"
	"github.com/stretchr/testify/assert"
)

// Tests only the jumpscriptInterceptor
func TestInterceptingPoppedCommands(t *testing.T) {

	jumpscriptStore := inmemorydata.NewJumpScriptStore()

	dummyContent := core.JumpScriptContent("My awesome script content")
	var command *core.Command
	{
		commandData := make(map[string]interface{})
		commandData["content"] = dummyContent

		rawCommand := make(core.RawCommand)
		rawCommand["data"] = string(utils.MustJsonMarshal(commandData))
		rawCommand["cmd"] = "jumpscript_content"

		var err error
		command, err = core.CommandFromRawCommand(rawCommand)
		assert.NoError(t, err)
	}


	commandSource := new(mocks.CommandSource)
	commandSource.On("BlockingPop").Return(command, nil)

	interceptedCommandSource := interceptors.NewInterceptedCommandSource(commandSource, jumpscriptStore)

	poppedCommand, err := interceptedCommandSource.BlockingPop()
	assert.NoError(t, err)

	poppedCommandData := make(map[string]interface{})
	utils.MustJsonUnmarshal([]byte(poppedCommand.Raw["data"].(string)), &poppedCommandData)

	hash := poppedCommandData["hash"]
	assert.NotNil(t, hash)

	id := core.JumpScriptID(hash.(string))

	retrievedContent, err := jumpscriptStore.Get(id)
	assert.NoError(t, err)

	assert.Equal(t, retrievedContent, dummyContent)
}