package rest

import (
	"github.com/Jumpscale/agentcontroller8/configs"
	"github.com/Jumpscale/agentcontroller8/core"
	"github.com/Jumpscale/agentcontroller8/events"
	"github.com/gin-gonic/gin"
)

type Manager struct {
	engine              *gin.Engine
	eventHandler        *events.Handler
	producerChanFactory core.ProducerChanFactory
	commandResponder    core.CommandResponder
	settings            *configs.Settings
	agentLog            core.AgentLog
	jumpscriptStore 	core.JumpScriptStore
}

func NewManager(eventHandler *events.Handler,
	producerChanFactory core.ProducerChanFactory,
	commandResponder core.CommandResponder,
	settings *configs.Settings,
	agentLog core.AgentLog,
	jumpscriptStore core.JumpScriptStore,
	) *Manager {

	r := Manager{
		engine:              gin.New(),
		eventHandler:        eventHandler,
		producerChanFactory: producerChanFactory,
		commandResponder:    commandResponder,
		settings:            settings,
		agentLog: 			 agentLog,
		jumpscriptStore: 	 jumpscriptStore,
	}

	r.engine.Use(gin.Recovery())
	r.engine.Use(LoggerWithWriter(gin.DefaultWriter))

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
