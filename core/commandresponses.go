package core

import (
	"encoding/json"
	"log"
	"time"
)

// Constructs an error response for the specified command
func ErrorResponseFor(command *Command, message string) *CommandResponse {
	content := CommandResponseContent{
		ID:        command.Content.ID,
		Gid:       command.Content.Gid,
		Nid:       command.Content.Nid,
		Tags:      command.Content.Tags,
		State:     CommandStateError,
		Data:      message,
		StartTime: int64(time.Duration(time.Now().UnixNano()) / time.Millisecond),
	}
	response := CommandResponseFromContent(&content)
	return response
}

// Constructs a response indicating that the command is queued on the given agent
func QueuedResponseFor(command *Command, queuedOn AgentID) *CommandResponse {
	content := CommandResponseContent{
		ID:        command.Content.ID,
		Gid:       int(queuedOn.GID),
		Nid:       int(queuedOn.NID),
		Tags:      command.Content.Tags,
		State:     CommandStateQueued,
		StartTime: int64(time.Duration(time.Now().UnixNano()) / time.Millisecond),
	}

	response := CommandResponseFromContent(&content)

	return response
}

// Constructs a response indicating that the given command is currently running on the specified agent
func RunningResponseFor(command *Command, runningOn AgentID) *CommandResponse {

	content := CommandResponseContent{
		ID:        command.Content.ID,
		Gid:       int(runningOn.GID),
		Nid:       int(runningOn.NID),
		Tags:      command.Content.Tags,
		State:     CommandStateRunning,
		StartTime: int64(time.Duration(time.Now().UnixNano()) / time.Millisecond),
	}

	response := CommandResponseFromContent(&content)

	return response
}

func SuccessResponseFor(command *Command, result interface{}, level int) *CommandResponse {

	jsonResult, err := json.Marshal(result)
	if err != nil {
		log.Fatalf("Failed to serialize: %v", err)
	}

	content := CommandResponseContent{
		ID:        command.Content.ID,
		Gid:       command.Content.Gid,
		Nid:       command.Content.Nid,
		Tags:      command.Content.Tags,
		State:     CommandStateSuccess,
		Data:      string(jsonResult),
		Level:     level,
		StartTime: int64(time.Duration(time.Now().UnixNano()) / time.Millisecond),
	}

	response := CommandResponseFromContent(&content)
	return response
}

func UnknownCommandResponseFor(command *Command) *CommandResponse {
	content := CommandResponseContent{
		ID:        command.Content.ID,
		Gid:       command.Content.Gid,
		Nid:       command.Content.Nid,
		Tags:      command.Content.Tags,
		State:     CommandStateErrorUnknownCommand,
		StartTime: int64(time.Duration(time.Now().UnixNano()) / time.Millisecond),
	}

	response := CommandResponseFromContent(&content)
	return response
}
