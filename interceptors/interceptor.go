// Middleware for received commands that kick in before received commands are dispatched or executed
package interceptors

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"github.com/garyburd/redigo/redis"
	"github.com/Jumpscale/agentcontroller2/messages"
)

const (
	scriptHashTimeout = 86400 // seconds
)

type commandInterceptor func(map[string]interface{}, *Manager) (map[string]interface{}, error)

type Manager struct {
	redisPool *redis.Pool
	interceptors map[string]commandInterceptor
}

func NewManager(redisPool *redis.Pool) *Manager {
	return &Manager {
		redisPool: redisPool,
		interceptors: map[string]commandInterceptor{
			"jumpscript_content": jumpscriptHasherInterceptor,
		},
	}
}

// Hashes jumpscripts executed by the jumpscript_content and store it in redis. Alters the passed command as needed
func jumpscriptHasherInterceptor(cmd map[string]interface{}, manager *Manager) (map[string]interface{}, error) {
	datastr, ok := cmd["data"].(string)
	if !ok {
		return nil, errors.New("Expecting command 'data' to be string")
	}

	data := make(map[string]interface{})
	err := json.Unmarshal([]byte(datastr), &data)
	if err != nil {
		return nil, err
	}

	content, ok := data["content"]
	if !ok {
		return nil, errors.New("jumpscript_content doesn't have content payload")
	}

	contentstr, ok := content.(string)
	if !ok {
		return nil, errors.New("Expected 'content' to be string")
	}

	hash := fmt.Sprintf("%x", md5.Sum([]byte(contentstr)))

	db := manager.redisPool.Get()
	defer db.Close()

	if _, err := db.Do("SET", hash, contentstr, "EX", scriptHashTimeout); err != nil {
		return nil, err
	}

	//hash is stored. Now modify the command and forward it.
	delete(data, "content")
	data["hash"] = hash

	updatedDatastr, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	cmd["data"] = string(updatedDatastr)

	return cmd, nil
}

func (manager *Manager) Intercept(command *messages.CommandMessage) *messages.CommandMessage {

	cmd := command.Raw

	cmdName, ok := cmd["cmd"].(string)
	if !ok {
		log.Println("Expected 'cmd' to be string")
		return command
	}

	interceptor, ok := manager.interceptors[cmdName]
	if !ok {
		return command
	}

	updatedRawCommand, err := interceptor(cmd, manager)
	if err != nil {
		log.Println("Failed to intercept command", err)
		return command
	}

	updatedCommand, err := messages.CommandMessageFromRawCommand(updatedRawCommand)
	if err != nil {
		log.Println(err)
		return command
	}

	return updatedCommand
}
