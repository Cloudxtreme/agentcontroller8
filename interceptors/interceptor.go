// Middleware for received commands that kick in before received commands are dispatched or executed
package interceptors

import (
	"log"
	"github.com/Jumpscale/agentcontroller8/core"
)

const (
	scriptHashTimeout = 86400 // seconds
)

type commandInterceptor func(core.RawCommand) (core.RawCommand, error)

type manager struct {
	jumpscriptStore core.JumpScriptStore
	interceptors    map[string]commandInterceptor
}

func newManager(jumpscriptStore core.JumpScriptStore) *manager {
	return &manager{
		jumpscriptStore: jumpscriptStore,
		interceptors: map[string]commandInterceptor{
			"jumpscript_content": jumpscriptInterceptor(jumpscriptStore),
		},
	}
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

	updatedRawCommand, err := interceptor(command.Raw)
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
