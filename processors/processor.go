//Processor is an extension to agentcontroller that does further processing on results and commands
//by calling external python code.
//The current processor impl will load the python module (defined by the config.Extension) and then call
//process_command for each received command and process_result for each received result.
package processors

import (
	"fmt"
	"github.com/Jumpscale/agentcontroller2/configs"
	"github.com/Jumpscale/agentcontroller2/core"
	"github.com/Jumpscale/pygo"
	"github.com/garyburd/redigo/redis"
	"log"
	"os"
)

type Processor interface {
	Start()
}

type processorImpl struct {
	enabled        bool
	commandResults core.LoggedCommandResults
	commands       core.LoggedCommands
	pool           *redis.Pool
	module         pygo.Pygo
}

//NewProcessor Creates a new processor
func NewProcessor(config *configs.Extension, pool *redis.Pool,
	commands core.LoggedCommands, commandResults core.LoggedCommandResults) (Processor, error) {

	var module pygo.Pygo
	var err error

	if config.Enabled {
		opts := &pygo.PyOpts{
			PythonPath: config.PythonPath,
			Env: []string{
				fmt.Sprintf("HOME=%s", os.Getenv("HOME")),
			},
		}

		module, err = pygo.NewPy(config.Module, opts)
		if err != nil {
			return nil, err
		}
	}

	processor := &processorImpl{
		enabled:       config.Enabled,
		pool:          pool,
		commandResults:  commandResults,
		commands: commands,
		module:        module,
	}

	return processor, nil
}

func (processor *processorImpl) processSingleResult() error {

	commandResultMessage, err := processor.commandResults.Pop()

	if err != nil {
		if core.IsTimeout(err) {
			return nil
		}

		return err
	}

	if processor.enabled {
		_, err := processor.module.Call("process_result", commandResultMessage.Content)
		if err != nil {
			log.Println("Processor", "Failed to process result", err)
		}
	}
	//else discard result

	return nil
}

func (processor *processorImpl) processSingleCommand() error {

	commandMessage, err := processor.commands.Pop()

	if err != nil {
		if core.IsTimeout(err) {
			return nil
		}

		return err
	}

	if processor.enabled {
		_, err := processor.module.Call("process_command", commandMessage.Raw)
		if err != nil {
			log.Println("Processor", "Failed to process command", err)
		}
	}
	//else discard command

	return nil
}

func (processor *processorImpl) resultsLoop() {
	for {
		err := processor.processSingleResult()
		if err != nil {
			log.Fatal("Processor", "Coulnd't read results from redis", err)
		}
	}
}

func (processor *processorImpl) commandsLoop() {
	for {
		err := processor.processSingleCommand()
		if err != nil {
			log.Fatal("Processor", "Coulnd't read commands from redis", err)
		}
	}
}

func (processor *processorImpl) Start() {
	go processor.resultsLoop()
	go processor.commandsLoop()
}
