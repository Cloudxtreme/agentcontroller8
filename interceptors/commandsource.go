package interceptors
import (
	"github.com/Jumpscale/agentcontroller2/core"
	"github.com/garyburd/redigo/redis"
)

type interceptedCommands struct {
	core.CommandSource
	interceptor *manager
}

func (interceptor *interceptedCommands) Pop() (*core.Command, error) {
	freshCommand, err := interceptor.CommandSource.Pop()
	if err != nil {
		return nil, err
	}
	mutatedCommand := interceptor.interceptor.Intercept(freshCommand)
	return mutatedCommand, nil
}

// Returns a core.IncomingCommands implementation that intercepts the commands received from the passed
// source of commands and mutates commands on-the-fly.
func NewInterceptedCommandSource(source core.CommandSource, redisConnPool *redis.Pool) core.CommandSource {
	return &interceptedCommands{
		CommandSource: source,
		interceptor: newManager(redisConnPool),
	}
}