// Post-execution processing of commands and command results
package commandprocessing

import (
	"fmt"
	"github.com/Jumpscale/agentcontroller8/configs"
	"github.com/Jumpscale/agentcontroller8/core"
	"github.com/Jumpscale/pygo"
	"github.com/garyburd/redigo/redis"
	"log"
	"os"
)

type CommandProcessor interface {
	Start()
}

type processorImpl struct {
	enabled        bool
	commandResults core.CommandResponseLog
	commands       core.CommandLog
	pool           *redis.Pool
	module         pygo.Pygo
}

//NewProcessor Creates a new processor
func NewProcessor(config *configs.Extension, pool *redis.Pool,
	commands core.CommandLog, commandResults core.CommandResponseLog) (CommandProcessor, error) {

	var module pygo.Pygo
	var err error

	if config.Enabled {
		opts := &pygo.PyOpts{
			PythonBinary: config.GetPythonBinary(),
			PythonPath:   config.PythonPath,
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
		enabled:        config.Enabled,
		pool:           pool,
		commandResults: commandResults,
		commands:       commands,
		module:         module,
	}

	return processor, nil
}

func (processor *processorImpl) processSingleResult() error {

	commandResultMessage, err := processor.commandResults.BlockingPop()

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

	commandMessage, err := processor.commands.BlockingPop()

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
