package core
import (
	"encoding/json"
)

const (
	CommandStateQueued  = "QUEUED"
	CommandStateRunning = "RUNNING"
	CommandStateError   = "ERROR"
	CommandStateSuccess = "SUCCESS"
	CommandStateErrorUnknownCommand = "UNKNOWN_CMD"
)

type CommandContent struct {
	ID     string   `json:"id"`
	Gid    int      `json:"gid"`
	Nid    int      `json:"nid"`
	Cmd    string   `json:"cmd"`
	Roles  []string `json:"roles"`
	Fanout bool     `json:"fanout"`
	Data   string   `json:"data"`
	Tags   string   `json:"tags"`
	Args   struct {
		Name string `json:"name"`
	} `json:"args"`
}

type RawCommand map[string]interface{}

type CommandResultContent struct {
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

type CommandResult struct {
	Content CommandResultContent
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
		JSON: payload,
		Raw: rawCommand,
	}, nil
}

func CommandFromRawCommand(rawCommand RawCommand) (*Command, error) {
	jsonData, err := json.Marshal(rawCommand)
	if err != nil {
		return nil, err
	}
	return CommandFromJSON(jsonData)
}

func CommandResultFromJSON(payload []byte) (*CommandResult, error) {
	var commandResult CommandResultContent
	err := json.Unmarshal(payload, &commandResult)
	if err != nil {
		return nil, err
	}

	return &CommandResult{
		Content: commandResult,
		JSON: payload,
	}, nil
}

func CommandResultFromCommandResultContent(commandResult *CommandResultContent) (*CommandResult, error) {
	jsonData, err := json.Marshal(commandResult)
	if err != nil {
		return nil, err
	}
	return CommandResultFromJSON(jsonData)
}

func (command *Command) String() string {
	return string(command.JSON)
}

func (command *CommandResult) String() string {
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