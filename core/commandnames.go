package core

// A non-exhaustive list of possible command names
const (
	CommandInternalListAgents CommandName = "list_agents"

	CommandInternalSchedulerAdd CommandName = "scheduler_add"
	CommandInternalSchedulerList CommandName = "scheduler_list"
	CommandInternalSchedulerRemove CommandName = "scheduler_remove"
	CommandInternalSchedulerRemovePrefix CommandName = "scheduler_remove_prefix"
)