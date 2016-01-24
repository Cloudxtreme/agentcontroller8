package core

import (
	"encoding/json"
	"log"
)

const (
	CommandStateQueued              = "QUEUED"
	CommandStateRunning             = "RUNNING"
	CommandStateError               = "ERROR"
	CommandStateSuccess             = "SUCCESS"
	CommandStateErrorUnknownCommand = "UNKNOWN_CMD"
)

type CommandName string

type CommandContent struct {
	ID     string      `json:"id"`
	Gid    int         `json:"gid"`
	Nid    int         `json:"nid"`
	Cmd    string      `json:"cmd"`
	Roles  []string    `json:"roles"`
	Fanout bool        `json:"fanout"`
	Data   string      `json:"data"`
	Tags   string      `json:"tags"`
	Args   CommandArgs `json:"args"`
}

type CommandArgs struct {
	Domain  string   `json:"domain"`
	Name    string   `json:"name"`
	Queue   string   `json:"queue"`
	MaxTime int      `json:"max_time"`
	Args    []string `json:"args"`
}

type RawCommand map[string]interface{}

type CommandResponseContent struct {
	ID        string                 `json:"id"`
	Gid       int                    `json:"gid"`
	Nid       int                    `json:"nid"`
	Cmd       string                 `json:"cmd"`
	Args      map[string]interface{} `json:"args"`
	Data      string                 `json:"data"`
	Streams   []string               `json:"streams"`
	Critical  string                 `json:"critical"`
	Tags      string                 `json:"tags"`
	Level     int                    `json:"level"`
	StartTime int64                  `json:"starttime"`
	State     string                 `json:"state"`
	Time      int                    `json:"time"`
}

type Command struct {
	Content CommandContent
	JSON    []byte
	Raw     RawCommand
}

type CommandResponse struct {
	Content CommandResponseContent
	JSON    []byte
}

func CommandFromJSON(payload []byte) (*Command, error) {
	var command CommandContent
	err := json.Unmarshal(payload, &command)
	if err != nil {
		return nil, err
	}

	var rawCommand RawCommand
	err = json.Unmarshal(payload, &rawCommand)
	if err != nil {
		return nil, err
	}

	return &Command{
		Content: command,
		JSON:    payload,
		Raw:     rawCommand,
	}, nil
}

func CommandFromRawCommand(rawCommand RawCommand) (*Command, error) {
	jsonData, err := json.Marshal(rawCommand)
	if err != nil {
		return nil, err
	}
	return CommandFromJSON(jsonData)
}

func CommandResponseFromJSON(payload []byte) (*CommandResponse, error) {
	var commandResult CommandResponseContent
	err := json.Unmarshal(payload, &commandResult)
	if err != nil {
		return nil, err
	}

	return &CommandResponse{
		Content: commandResult,
		JSON:    payload,
	}, nil
}

func CommandResponseFromContent(content *CommandResponseContent) *CommandResponse {
	jsonData, err := json.Marshal(content)
	if err != nil {
		log.Fatal(err)
	}
	return &CommandResponse{
		Content: *content,
		JSON:    jsonData,
	}
}

func CommandFromContent(content *CommandContent) *Command {
	jsonData, err := json.Marshal(content)
	if err != nil {
		log.Fatal(err)
	}
	command, err := CommandFromJSON(jsonData)
	if err != nil {
		log.Fatal(err)
	}
	return command
}

func (command *Command) String() string {
	return string(command.JSON)
}

func (command *CommandResponse) String() string {
	return string(command.JSON)
}

func (command *Command) IsInternal() bool {
	return command.Content.Cmd == "controller"
}

func (command *Command) AttachedRoles() []AgentRole {
	var roles []AgentRole
	for _, role := range command.Content.Roles {
		roles = append(roles, AgentRole(role))
	}
	return roles
}

// Returns nil if no GID was attached
func (command *Command) AttachedGID() *uint {
	if command.Content.Gid == 0 {
		return nil
	}
	gid := uint(command.Content.Gid)
	return &gid
}

func IsTerminalCommandState(state string) bool {
	return state != CommandStateQueued && state != CommandStateRunning
}
