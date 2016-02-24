package interceptors
import (
	"github.com/Jumpscale/agentcontroller8/core"
)

type interceptedCommands struct {
	core.CommandSource
	i *manager
}

func (interceptor *interceptedCommands) BlockingPop() (*core.Command, error) {
	freshCommand, err := interceptor.CommandSource.BlockingPop()
	if err != nil {
		return nil, err
	}
	mutatedCommand := interceptor.i.Intercept(freshCommand)
	return mutatedCommand, nil
}

// Returns a core.IncomingCommands implementation that intercepts the commands received from the passed
// source of commands and mutates commands on-the-fly.
func NewInterceptedCommandSource(source core.CommandSource, jumpscriptStore core.JumpScriptStore) core.CommandSource {
	return &interceptedCommands{
		CommandSource: source,
		i: newManager(jumpscriptStore),
	}
}