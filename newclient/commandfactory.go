package newclient
import (
	"github.com/Jumpscale/agentcontroller2/core"
	"github.com/pborman/uuid"
)

type CommandTarget struct {

	// Target grid ID, must be nonzero
	GID    uint

	// Target node ID, must be nonzero
	NID    uint

	// Target roles
	Roles  []core.AgentRole

	// When Fanout is true all matching agents are targeted, otherwise
	// an arbitrary agent that matches is targeted
	Fanout bool
}


// Command-specific argument
type CommandArguments struct {

	Name       string

	Domain     string

	// Internal queue on Agent for job execution
	Queue      string

	// Maximum time allowed for the command to execute (0 is forever)
	MaxRunTime uint
}

type CommandFactory struct {

	Target    CommandTarget

	Name      core.CommandName

	// Arbitrary labels for commands
	Tags      []string

	// Command-specific data in an a format expected by the command executor
	Data      string

	Arguments CommandArguments
}

func (target CommandTarget) AddTargetRole(role core.AgentRole) {
	target.Roles = append(target.Roles, role)
}

func (factory CommandFactory) AddTag(tag string) {
	factory.Tags = append(factory.Tags, tag)
}

func (factory CommandFactory) Build() *core.Command {

	var roles []string
	for _, role := range factory.Target.Roles {
		roles = append(roles, string(role))
	}

	content := &core.CommandContent{
		ID: uuid.New(),
		Gid: int(factory.Target.GID),
		Nid: int(factory.Target.NID),
		Roles: roles,
		Fanout: factory.Target.Fanout,

		Cmd: string(factory.Name),
		Data: factory.Data,
		Tags: factory.Data,
		Args: core.CommandArgs {
			Name: factory.Arguments.Name,
			Queue: factory.Arguments.Queue,
			MaxTime: int(factory.Arguments.MaxRunTime),
			Domain: factory.Arguments.Domain,
		},
	}

	return core.CommandFromContent(content)
}


// Builds and returns a GetProcessStats command for the given target
func CommandGetProcessStats(target CommandTarget) *core.Command {
	return CommandFactory{
		Name: core.CommandGetProcessStats,
		Target: target,
	}.Build()
}


func CommandInternalListAgents() *core.Command {
	return CommandFactory{
		Name: core.CommandInternalListAgents,
	}.Build()
}


func CommandInternalSchedulerListJobs() *core.Command {
	return CommandFactory{
		Name: core.CommandInternalSchedulerListJobs,
	}.Build()
}
