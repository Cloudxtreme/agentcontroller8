package redisdata
import "github.com/Jumpscale/agentcontroller2/core"

// A CommandSource that logs about each popped command in its internal log
type LoggedCommandSource struct {
	core.CommandSource
	Log core.CommandLog
}

func (commandSource *LoggedCommandSource) Pop() (*core.Command, error) {
	command, err := commandSource.CommandSource.Pop()
	if err != nil {
		return nil, err
	}
	commandSource.Log.Push(command)
	return command, err
}
