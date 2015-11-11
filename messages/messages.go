// Data structures for incoming and outgoing messages in multiple formats, and other data types that make use of them.
package messages
import (
	"github.com/Jumpscale/agentcontroller2/core"
	"encoding/json"
)

// A core.Command in multiple formats.
// The payload is the JSON encoding of the content (it may contain more data than represented in the command)
type CommandMessage struct {
	Content core.Command
	Payload []byte
	Raw		core.RawCommand
}

// A core.CommandResult in multiple formats.
// The payload is the JSON encoding of the content (it may contain more data than represented in the command result)
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

func CommandResultMessageFromJSON(payload []byte) (*CommandResultMessage, error) {
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

func CommandResultMessageFromCommandResult(commandResult *core.CommandResult) (*CommandResultMessage, error) {
	jsonData, err := json.Marshal(commandResult)
	if err != nil {
		return nil, err
	}
	return CommandResultMessageFromJSON(jsonData)
}

func (message *CommandMessage) String() string {
	return string(message.Payload)
}

func (message *CommandResultMessage) String() string {
	return string(message.Payload)
}