package rest
import (
	"github.com/gin-gonic/gin"
	"github.com/Jumpscale/agentcontroller2/happenings"
	"github.com/Jumpscale/agentcontroller2/core"
	"github.com/garyburd/redigo/redis"
	"github.com/Jumpscale/agentcontroller2/configs"
)

type RestInterface struct {
	engine *gin.Engine
	eventHandler *happenings.EventsHandler
	producerChanFactory core.ProducerChanFactory
	redisPool *redis.Pool
	commandResponder core.CommandResponder
	settings 	*configs.Settings
}

func NewRestInterface(
	eventHandler *happenings.EventsHandler,
	producerChanFactory core.ProducerChanFactory,
	redisPool *redis.Pool,
	commandResponder core.CommandResponder,
	settings *configs.Settings,
	) *RestInterface {

	r := RestInterface{
		engine: gin.Default(),
		eventHandler: eventHandler,
		producerChanFactory: producerChanFactory,
		redisPool: redisPool,
		commandResponder: commandResponder,
		settings: settings,
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

func (r *RestInterface) Engine() *gin.Engine {
	return r.engine
}