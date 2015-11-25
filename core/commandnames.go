package core

// A non-exhaustive list of possible command names
const (

	CommandInternalListAgents CommandName = "list_agents"

	CommandInternalSchedulerAddJob CommandName = "scheduler_add"
	CommandInternalSchedulerListJobs CommandName = "scheduler_list"
	CommandInternalSchedulerRemoveJob CommandName = "scheduler_remove"
	CommandInternalSchedulerRemoveJobByIdPrefix CommandName = "scheduler_remove_prefix"

	CommandGetProcessStats CommandName = "get_processes_stats"
)