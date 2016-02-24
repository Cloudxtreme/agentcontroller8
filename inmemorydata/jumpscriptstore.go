package inmemorydata
import (
	"github.com/Jumpscale/agentcontroller8/core"
	"sync"
	"fmt"
	"github.com/Jumpscale/agentcontroller8/utils"
)

type jumpscriptStore struct {
	mapping map[core.JumpScriptID]core.JumpScriptContent
	lock    sync.RWMutex
}

func NewJumpScriptStore() core.JumpScriptStore {
	return &jumpscriptStore{
		mapping: make(map[core.JumpScriptID]core.JumpScriptContent),
	}
}


func (store *jumpscriptStore) Add(content core.JumpScriptContent) (core.JumpScriptID, error) {
	store.lock.Lock()
	defer store.lock.Unlock()

	id := core.JumpScriptID(utils.MD5Hex([]byte(content)))
	store.mapping[id] = content

	return id, nil
}

func (store *jumpscriptStore) Get(id core.JumpScriptID) (core.JumpScriptContent, error) {
	store.lock.RLock()
	defer store.lock.RUnlock()

	content, exists := store.mapping[id]
	if !exists {
		return core.JumpScriptContent(""), fmt.Errorf("ID does not exist")
	}

	return content, nil
}