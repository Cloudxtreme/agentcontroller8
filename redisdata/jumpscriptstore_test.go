package redisdata_test
import (
	"testing"
	"github.com/Jumpscale/agentcontroller8/redisdata"
	"github.com/stretchr/testify/assert"
	"github.com/Jumpscale/agentcontroller8/core"
)

func TestJumpScriptStore(t *testing.T) {
	pool := TestingRedisPool(t)
	store := redisdata.NewJumpScriptStore(pool)

	content := core.JumpScriptContent("DUMMY script ConTENT")

	id, err := store.Add(content)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	retrievedContent, err := store.Get(id)
	assert.NoError(t, err)
	assert.Equal(t, content, retrievedContent)
}