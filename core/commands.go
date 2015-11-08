package core

const (
	COMMAND_STATE_QUEUED = "QUEUED"
	COMMAND_STATE_RUNNING = "RUNNING"
	COMMAND_STATE_ERROR	= "ERROR"
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

type CommandResult struct {
	ID        string   `json:"id"`
	Gid       int      `json:"gid"`
	Nid       int      `json:"nid"`
	Cmd       string   `json:"cmd"`
	Data      string   `json:"data"`
	Streams   []string `json:"streams"`
	Tags      string   `json:"tags"`
	Level     int      `json:"level"`
	StartTime int64    `json:"starttime"`
	State     string   `json:"state"`
	Time      int      `json:"time"`
}