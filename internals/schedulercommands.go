package internals
import (
"github.com/Jumpscale/agentcontroller2/scheduling"
"github.com/Jumpscale/agentcontroller2/core"
"encoding/json"
)

const (

)

func (manager *Manager) setUpSchedulerCommands(scheduler *scheduling.Scheduler) {

	manager.commandHandlers[SchedulerAddJob] =
		func(cmd *core.Command) (interface{}, error) {
			job, err := scheduling.JobFromJSON([]byte(cmd.Content.Data))
			if err != nil {
				return nil, err
			}
			return nil, scheduler.AddJob(job)
		}

	manager.commandHandlers[SchedulerListJobs] =
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

	manager.commandHandlers[SchedulerRemoveJob] =
		func (cmd *core.Command) (interface{}, error) {
			return scheduler.RemoveByID(cmd.Content.ID)
		}

	manager.commandHandlers[SchedulerRemoveJobByIdPrefix] =
		func (cmd *core.Command) (interface{}, error) {
			scheduler.RemoveByIdPrefix(cmd.Content.ID)
			return nil, nil
		}
}