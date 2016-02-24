package logged
import (
	"github.com/Jumpscale/agentcontroller8/core"
	"fmt"
)

// A CommandResponder that logs its responses in its internal log
type CommandResponder struct {
	core.CommandResponder
	Log core.CommandResponseLog
}

func (responder *CommandResponder) RespondToCommand(response *core.CommandResponse) error {
	err := responder.CommandResponder.RespondToCommand(response)
	if err != nil {
		return err
	}
	err = responder.Log.Push(response)
	if err != nil {
		return fmt.Errorf("Failed to log the response: %v", err)
	}

	return nil
}

