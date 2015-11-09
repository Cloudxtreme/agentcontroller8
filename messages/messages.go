package messages
import (
	"github.com/Jumpscale/agentcontroller2/core"
	"encoding/json"
)

// A core.Command in transit. The payload is the JSON encoding of the content (it may contain
// more data than represented in the command)
type CommandMessage struct {
	Content core.Command
	Payload []byte
	Raw		core.RawCommand
}

// A core.CommandResult in transit. The payload is the JSON encoding of the content (it may contain
// more data than represented in the command result)
type CommandResultMessage struct {
	Content core.CommandResult
	Payload []byte
}

func CommandMessageFromJSON(payload []byte) (*CommandMessage, error) {
	var command core.Command
	err := json.Unmarshal(payload, &command)
	if err != nil {
		return nil, err
	}

	var rawCommand core.RawCommand
	err = json.Unmarshal(payload, &rawCommand)
	if err != nil {
		return nil, err
	}

	return &CommandMessage{
		Content: command,
		Payload: payload,
		Raw: rawCommand,
	}, nil
}

func CommandMessageFromRawCommand(rawCommand core.RawCommand) (*CommandMessage, error) {
	jsonData, err := json.Marshal(rawCommand)
	if err != nil {
		return nil, err
	}
	return CommandMessageFromJSON(jsonData)
}

func CommandResultMessageFrom(payload []byte) (*CommandResultMessage, error) {
	var commandResult core.CommandResult
	err := json.Unmarshal(payload, &commandResult)
	if err != nil {
		return nil, err
	}

	return &CommandResultMessage{
		Content: commandResult,
		Payload: payload,
	}, nil
}