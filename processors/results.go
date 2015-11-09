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

type Processor interface {
	Start()
}

type resultsProcessorImpl struct {
	enabled bool
	queue   messages.RedisCommandResultList
	pool    *redis.Pool

	module pygo.Pygo
}

func NewResultsProcessor(config *configs.Extension, pool *redis.Pool,
	queue messages.RedisCommandResultList) (Processor, error) {

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

	processor := &resultsProcessorImpl{
		enabled: config.Enabled,
		pool:    pool,
		queue:   queue,
		module:  module,
	}

	return processor, nil
}

func (processor *resultsProcessorImpl) processSingleResult() error {

	commandResultMessage, err := processor.queue.BlockingPop(processor.pool, 0)

	if err != nil {
		if core.IsTimeout(err) {
			return nil
		}

		return err
	}

	if processor.enabled {
		_, err := processor.module.Call("process", commandResultMessage.Content)
		if err != nil {
			log.Println("Failed to process result", err)
		}
	}
	//else discard command

	return nil
}

func (processor *resultsProcessorImpl) loop() {
	for {
		err := processor.processSingleResult()
		if err != nil {
			log.Fatal("Coulnd't read results from redis", err)
		}
	}
}

func (processor *resultsProcessorImpl) Start() {
	go processor.loop()
}