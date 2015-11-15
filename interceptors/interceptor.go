// Middleware for received commands that kick in before received commands are dispatched or executed
package interceptors

import (
	"encoding/json"
	"errors"
	"log"
	"github.com/Jumpscale/agentcontroller2/core"
)

const (
	scriptHashTimeout = 86400 // seconds
)

type commandInterceptor func(map[string]interface{}, *manager) (map[string]interface{}, error)

type manager struct {
	jumpscriptStore core.JumpScriptStore
	interceptors    map[string]commandInterceptor
}

func newManager(jumpscriptStore core.JumpScriptStore) *manager {
	return &manager{
		jumpscriptStore: jumpscriptStore,
		interceptors: map[string]commandInterceptor{
			"jumpscript_content": jumpscriptHasherInterceptor,
		},
	}
}

// Hashes jumpscripts executed by the jumpscript_content and store it in redis. Alters the passed command as needed
func jumpscriptHasherInterceptor(cmd map[string]interface{}, manager *manager) (map[string]interface{}, error) {
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

	jumpscriptContent, ok := content.(string)
	if !ok {
		return nil, errors.New("Expected 'content' to be string")
	}

	id, err := manager.jumpscriptStore.Add(core.JumpScriptContent(jumpscriptContent))

	if err != nil {
		return nil, err
	}

	//hash is stored. Now modify the command and forward it.
	delete(data, "content")
	data["hash"] = id

	updatedDatastr, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	cmd["data"] = string(updatedDatastr)

	return cmd, nil
}

func (manager *manager) Intercept(command *core.Command) *core.Command {

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

	updatedCommand, err := core.CommandFromRawCommand(updatedRawCommand)
	if err != nil {
		log.Println(err)
		return command
	}

	return updatedCommand
}
