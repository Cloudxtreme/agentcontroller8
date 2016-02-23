package logged
import (
	"github.com/Jumpscale/agentcontroller8/core"
	"fmt"
)

// A CommandSource that logs about each popped command in its internal log
type CommandSource struct {
	core.CommandSource
	Log core.CommandLog
}

func (commandSource *CommandSource) BlockingPop() (*core.Command, error) {
	command, err := commandSource.CommandSource.BlockingPop()
	if err != nil {
		return nil, err
	}
	err = commandSource.Log.Push(command)
	if err != nil {
		return command, fmt.Errorf("Failed to log the command: %v", err)
	}
	return command, err
}
