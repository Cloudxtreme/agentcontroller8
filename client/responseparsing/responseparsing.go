package responseparsing

import (
	"encoding/json"
	"fmt"
	"github.com/Jumpscale/agentcontroller8/core"
	"github.com/Jumpscale/agentcontroller8/scheduling"
	"github.com/Jumpscale/agentcontroller8/utils"
	"strings"
)

type ExecutableResult struct {
	StandardOut string
	StandardErr string
}

func InternalListAgents(response *core.CommandResponse) []core.AgentID {
	agentMap := map[string](interface{}){}
	err := json.Unmarshal([]byte(response.Content.Data), &agentMap)
	if err != nil {
		panic(fmt.Errorf("Malformed response"))
	}

	agents := []core.AgentID{}
	for agentStr, _ := range agentMap {
		gidnid := strings.Split(agentStr, ":")
		if len(gidnid) != 2 {
			panic("Malformed response")
		}
		agents = append(agents, utils.AgentIDFromStrings(gidnid[0], gidnid[1]))
	}

	return agents
}

func Execute(response *core.CommandResponse) ExecutableResult {
	return ExecutableResult{
		StandardOut: response.Content.Streams[0],
		StandardErr: response.Content.Streams[1],
	}
}

type RunningCommandStats struct {
	Command struct {
		Args struct {
			Args          []string `json:"args"`
			Domain        string   `json:"domain"`
			MaxTime       int      `json:"max_time"`
			Name          string   `json:"name"`
			Queue         string   `json:"queue"`
			StatsInterval int      `json:"stats_interval"`
		} `json:"args"`
		Cmd   string   `json:"cmd"`
		Data  string   `json:"data"`
		GID   int      `json:"gid"`
		ID    string   `json:"id"`
		NID   int      `json:"nid"`
		Roles []string `json:"roles"`
		Tags  string   `json:"tags"`
	} `json:"cmd"`
	CPU   float64 `json:"cpu"`
	Debug string  `json:"debug"`
	Rss   int     `json:"rss"`
	Swap  int     `json:"swap"`
	Vms   int     `json:"vms"`
}

func GetProcessStats(response *core.CommandResponse) []RunningCommandStats {
	runningStats := []RunningCommandStats{}
	err := json.Unmarshal([]byte(response.Content.Data), &runningStats)
	if err != nil {
		panic(fmt.Errorf("Malformed response: %v in %v", err, string(response.JSON)))
	}
	return runningStats
}

func InternalSchedulerListJobs(response *core.CommandResponse) []scheduling.Job {
	jobs := []scheduling.Job{}
	err := json.Unmarshal([]byte(response.Content.Data), &jobs)
	if err != nil {
		panic(fmt.Errorf("Malformed response: %v", err))
	}
  if jobs == nil {
    // Because AC2 sometimes returns the literal "null" and this function should not be returning nils
    return []scheduling.Job{}
  }
	return jobs
}

func InternalSchedulerRemoveJob(response *core.CommandResponse) bool {
	return string(response.Content.Data) == "1"
}
