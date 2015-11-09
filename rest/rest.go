package rest

import (
	"github.com/Jumpscale/agentcontroller2/configs"
	"github.com/Jumpscale/agentcontroller2/core"
	"github.com/Jumpscale/agentcontroller2/events"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
)

type Manager struct {
	engine              *gin.Engine
	eventHandler        *events.EventsHandler
	producerChanFactory core.ProducerChanFactory
	redisPool           *redis.Pool
	commandResponder    core.CommandResponder
	settings            *configs.Settings
}

func NewManager(
	eventHandler *events.EventsHandler,
	producerChanFactory core.ProducerChanFactory,
	redisPool *redis.Pool,
	commandResponder core.CommandResponder,
	settings *configs.Settings,
) *Manager {

	r := Manager{
		engine:              gin.Default(),
		eventHandler:        eventHandler,
		producerChanFactory: producerChanFactory,
		redisPool:           redisPool,
		commandResponder:    commandResponder,
		settings:            settings,
	}

	r.engine.GET("/:gid/:nid/cmd", r.cmd)
	r.engine.POST("/:gid/:nid/log", r.logs)
	r.engine.POST("/:gid/:nid/result", r.result)
	r.engine.POST("/:gid/:nid/stats", r.stats)
	r.engine.POST("/:gid/:nid/event", eventHandler.Event)
	r.engine.GET("/:gid/:nid/hubble", r.handlHubbleProxy)
	r.engine.GET("/:gid/:nid/script", r.script)
	// router.Static("/doc", "./doc")

	return &r
}

func (r *Manager) Engine() *gin.Engine {
	return r.engine
}
