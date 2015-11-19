package redisdata
import (
	"github.com/garyburd/redigo/redis"
	"github.com/Jumpscale/agentcontroller2/core"
	"crypto/md5"
	"fmt"
	"github.com/Jumpscale/agentcontroller2/redisdata/ds"
	"time"
)

type jumpscriptStore struct {
	connPool *redis.Pool
}

func NewJumpScriptStore(connPool *redis.Pool) core.JumpScriptStore {
	return &jumpscriptStore{
		connPool: connPool,
	}
}

const scriptTTL = 86400 * time.Second

func (store *jumpscriptStore) Add(content core.JumpScriptContent) (core.JumpScriptID, error) {

	id := fmt.Sprintf("%x", md5.Sum([]byte(content)))
	jumpscriptID := core.JumpScriptID(id)

	redisValue := ds.Value{Name: fmt.Sprintf("jumpscript:%s", id)}
	err := redisValue.Set(store.connPool, []byte(content))
	if err != nil {
		return jumpscriptID, err
	}

	err = redisValue.Expire(store.connPool, scriptTTL)
	if err != nil {
		return jumpscriptID, err
	}

	return jumpscriptID, nil
}

func (store *jumpscriptStore) Get(id core.JumpScriptID) (core.JumpScriptContent, error) {
	redisValue := ds.Value{Name: fmt.Sprintf("jumpscript:%s", string(id))}
	content, err := redisValue.Get(store.connPool)
	return core.JumpScriptContent(content), err
}
