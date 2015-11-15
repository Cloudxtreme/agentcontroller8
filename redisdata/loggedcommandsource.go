package redisdata
import (
	"github.com/Jumpscale/agentcontroller2/core"
	"fmt"
)

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
	err = commandSource.Log.Push(command)
	if err != nil {
		return command, fmt.Errorf("Failed to log the command: %v", err)
	}
	return command, err
}
