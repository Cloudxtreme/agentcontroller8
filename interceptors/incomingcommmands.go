package interceptors
import (
	"github.com/Jumpscale/agentcontroller2/core"
	"github.com/garyburd/redigo/redis"
)

type interceptedCommands struct {
	source      core.IncomingCommands
	interceptor *manager
}

func (interceptor *interceptedCommands) Pop() (*core.Command, error) {
	freshCommand, err := interceptor.source.Pop()
	if err != nil {
		return nil, err
	}
	mutatedCommand := interceptor.interceptor.Intercept(freshCommand)
	return mutatedCommand, nil
}

func (interceptor *interceptedCommands) Push(command *core.Command) error {
	return interceptor.source.Push(command)
}

// Returns a core.IncomingCommands implementation that intercepts the commands received from the passed
// source of commands and mutates commands on-the-fly.
func Intercept(source core.IncomingCommands, redisConnPool *redis.Pool) core.IncomingCommands {
	return &interceptedCommands{
		source: source,
		interceptor: newManager(redisConnPool),
	}
}