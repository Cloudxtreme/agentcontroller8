package application
import (
	"github.com/Jumpscale/agentcontroller2/core"
	"time"
)

func errorResponseFor(command *core.Command, message string) *core.CommandResponse {
	content := core.CommandReponseContent{
		ID:        command.Content.ID,
		Gid:       command.Content.Gid,
		Nid:       command.Content.Nid,
		Tags:      command.Content.Tags,
		State:     core.CommandStateError,
		Data:      message,
		StartTime: int64(time.Duration(time.Now().UnixNano()) / time.Millisecond),
	}
	response := core.CommandResponseFromContent(&content)
	return response
}

func queuedResponseFor(command *core.Command, queuedOn core.AgentID) *core.CommandResponse {
	content := core.CommandReponseContent{
		ID:        command.Content.ID,
		Gid:       int(queuedOn.GID),
		Nid:       int(queuedOn.NID),
		Tags:      command.Content.Tags,
		State:     core.CommandStateQueued,
		StartTime: int64(time.Duration(time.Now().UnixNano()) / time.Millisecond),
	}

	response := core.CommandResponseFromContent(&content)

	return response
}

func runningResponseFor(command *core.Command, runningOn core.AgentID) *core.CommandResponse {

	content := core.CommandReponseContent{
		ID:        command.Content.ID,
		Gid:       int(runningOn.GID),
		Nid:       int(runningOn.NID),
		Tags:      command.Content.Tags,
		State:     core.CommandStateRunning,
		StartTime: int64(time.Duration(time.Now().UnixNano()) / time.Millisecond),
	}

	response := core.CommandResponseFromContent(&content)

	return response
}