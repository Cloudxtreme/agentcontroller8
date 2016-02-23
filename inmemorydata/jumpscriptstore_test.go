package inmemorydata_test
import (
	"testing"
	"github.com/Jumpscale/agentcontroller8/inmemorydata"
	"github.com/Jumpscale/agentcontroller8/core"
	"github.com/stretchr/testify/assert"
)

func TestJumpScriptStoreTest(t *testing.T) {
	store := inmemorydata.NewJumpScriptStore()

	content := core.JumpScriptContent("My jumpscript CODE")

	id, err  := store.Add(content)
	assert.NoError(t, err)

	retrieved, err := store.Get(id)
	assert.NoError(t, err)

	assert.Equal(t, content, retrieved)
}