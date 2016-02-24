package interceptors
import (
	"testing"
	"github.com/Jumpscale/agentcontroller8/inmemorydata"
	"github.com/Jumpscale/agentcontroller8/core"
	"github.com/Jumpscale/agentcontroller8/utils"
	"github.com/stretchr/testify/assert"
)


func TestJumpScriptInterceptor(t *testing.T) {

	jumpscriptStore := inmemorydata.NewJumpScriptStore()

	interceptor := jumpscriptInterceptor(jumpscriptStore)

	dummyContent := core.JumpScriptContent("My awesome script content")

	commandData := make(map[string]interface{})
	commandData["content"] = dummyContent

	rawCommand := make(core.RawCommand)
	rawCommand["data"] = string(utils.MustJsonMarshal(commandData))

	mutilatedRawCommand, err := interceptor(rawCommand)
	assert.NoError(t, err)

	mutilatedCommandData := make(map[string]interface{})
	utils.MustJsonUnmarshal([]byte(mutilatedRawCommand["data"].(string)), &mutilatedCommandData)

	id := mutilatedCommandData["hash"].(string)

	retrieved, err := jumpscriptStore.Get(core.JumpScriptID(id))
	assert.NoError(t, err)

	assert.Equal(t, dummyContent, retrieved)
}