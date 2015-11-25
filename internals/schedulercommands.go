package internals
import (
"github.com/Jumpscale/agentcontroller2/scheduling"
"github.com/Jumpscale/agentcontroller2/core"
"encoding/json"
)

const (
	COMMAND_INTERNAL_SCHEDULER_ADD CommandName = "scheduler_add"
	COMMAND_INTERNAL_SCHEDULER_LIST CommandName = "scheduler_list"
	COMMAND_INTERNAL_SCHEDULER_REMOVE CommandName = "scheduler_remove"
	COMMAND_INTERNAL_SCHEDULER_REMOVE_PREFIX CommandName = "scheduler_remove"
)

func (manager *Manager) setUpSchedulerCommands(scheduler *scheduling.Scheduler) {

	manager.commandHandlers[COMMAND_INTERNAL_SCHEDULER_ADD] =
		func(cmd *core.Command) (interface{}, error) {
			job, err := scheduling.JobFromJSON([]byte(cmd.Content.Data))
			if err != nil {
				return nil, err
			}
			return nil, scheduler.AddJob(job)
		}

	manager.commandHandlers[COMMAND_INTERNAL_SCHEDULER_LIST] =
		func(_ *core.Command) (interface{}, error) {
			jobs := scheduler.ListJobs()
			jobsMap := make(map[string]string)
			for _, job := range jobs {
				jsonJob, err := json.Marshal(job)
				if err != nil {
					panic(err)
				}
				jobsMap[job.ID] = string(jsonJob)
			}
			return scheduler.ListJobs(), nil
		}

	manager.commandHandlers[COMMAND_INTERNAL_SCHEDULER_REMOVE] =
		func (cmd *core.Command) (interface{}, error) {
			return scheduler.RemoveByID(cmd.Content.ID)
		}

	manager.commandHandlers[COMMAND_INTERNAL_SCHEDULER_REMOVE_PREFIX] =
		func (cmd *core.Command) (interface{}, error) {
			scheduler.RemoveByIdPrefix(cmd.Content.ID)
			return nil, nil
		}
}