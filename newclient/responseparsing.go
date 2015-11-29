package newclient
import (
	"github.com/Jumpscale/agentcontroller2/core"
	"encoding/json"
	"fmt"
	"strings"
	"github.com/Jumpscale/agentcontroller2/utils"
)

type ExecutableResult struct {
	StandardOut string
	StandardErr string
}

func parseCommandInternalListAgents(response *core.CommandResponse) []core.AgentID {
	agentMap := map[string](interface{}){}
	err := json.Unmarshal([]byte(response.Content.Data), &agentMap)
	if err != nil {
		panic(fmt.Errorf("Malformed response"))
	}

	var agents []core.AgentID
	for agentStr, _ := range agentMap {
		gidnid := strings.Split(agentStr, ":")
		if len(gidnid) != 2 {
			panic("Malformed response")
		}
		agents = append(agents, utils.AgentIDFromStrings(gidnid[0], gidnid[1]))
	}

	return agents
}

func parseCommandExecute(response *core.CommandResponse) ExecutableResult {
	return ExecutableResult{
		StandardOut: response.Content.Streams[0],
		StandardErr: response.Content.Streams[1],
	}
}