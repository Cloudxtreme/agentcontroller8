package commandfactory
import (
	"github.com/Jumpscale/agentcontroller2/core"
	"github.com/Jumpscale/agentcontroller2/internals"
)

// Builds and returns a GetProcessStats command for the given target
func CommandGetProcessStats(target CommandTarget) *core.Command {
	return CommandFactory{
		Name: core.CommandGetProcessStats,
		Target: target,
		Data: "{\"domain\": null, \"name\": null}",		// Filtering is not supported in this client for simplicity
	}.Build()
}


func CommandInternalListAgents() *core.Command {
	return CommandFactory{
		Name: core.CommandInternal,
		Arguments: CommandArguments{
			Name: string(internals.ListAgents),
		},
	}.Build()
}


func CommandInternalSchedulerListJobs() *core.Command {
	return CommandFactory{
		Name: core.CommandInternal,
		Arguments: CommandArguments{
			Name: string(internals.SchedulerListJobs),
		},
	}.Build()
}

// Executes an executable on an Agent
func CommandExecute(target CommandTarget, executable string, args []string) *core.Command {
	return CommandFactory{
		Name: core.CommandExecute,
		Target: target,
		Arguments: CommandArguments{
			Name: executable,
			ExecutableArguments: args,
		},
	}.Build()
}