package processors

import (
	"fmt"
	"github.com/Jumpscale/agentcontroller2/configs"
	"github.com/Jumpscale/agentcontroller2/core"
	"github.com/Jumpscale/agentcontroller2/messages"
	"github.com/Jumpscale/pygo"
	"github.com/garyburd/redigo/redis"
	"log"
	"os"
)

type commandProcessorImpl struct {
	enabled bool
	queue   messages.RedisCommandList
	pool    *redis.Pool

	module pygo.Pygo
}

func NewCommandProcessor(config *configs.Extension, pool *redis.Pool,
	queue messages.RedisCommandList) (Processor, error) {

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

	processor := &commandProcessorImpl{
		enabled: config.Enabled,
		pool:    pool,
		queue:   queue,
		module:  module,
	}

	return processor, nil
}

func (processor *commandProcessorImpl) processSingleResult() error {

	commandMessage, err := processor.queue.BlockingPop(processor.pool, 0)

	if err != nil {
		if core.IsTimeout(err) {
			return nil
		}

		return err
	}

	if processor.enabled {
		_, err := processor.module.Call("process", commandMessage.Content)
		if err != nil {
			log.Println("Processor", "Failed to process command", err)
		}
	}
	//else discard command

	return nil
}

func (processor *commandProcessorImpl) loop() {
	for {
		err := processor.processSingleResult()
		if err != nil {
			log.Fatal("Processor", "Coulnd't read commands from redis", err)
		}
	}
}

func (processor *commandProcessorImpl) Start() {
	go processor.loop()
}
