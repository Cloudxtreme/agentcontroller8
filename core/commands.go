package core

const (
	CommandStateQueued  = "QUEUED"
	CommandStateRunning = "RUNNING"
	CommandStateError   = "ERROR"
	CommandStateSuccess = "SUCCESS"
	CommandStateErrorUnknownCommand = "UNKNOWN_CMD"
)

type Command struct {
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

type CommandResult struct {
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


