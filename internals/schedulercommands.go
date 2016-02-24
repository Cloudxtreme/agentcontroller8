package internals
import (
	"github.com/Jumpscale/agentcontroller8/scheduling"
	"github.com/Jumpscale/agentcontroller8/core"
	"encoding/json"
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
			return scheduler.ListJobs(), nil
		}

	manager.commandHandlers[SchedulerRemoveJob] =
		func (cmd *core.Command) (interface{}, error) {
			var jobID string
			err := json.Unmarshal([]byte(cmd.Content.Data), &jobID)
			if err != nil {
				return nil, err
			}
			return scheduler.RemoveByID(jobID)
		}

	manager.commandHandlers[SchedulerRemoveJobByIdPrefix] =
		func (cmd *core.Command) (interface{}, error) {
			var jobIDPrefix string
			err := json.Unmarshal([]byte(cmd.Content.Data), &jobIDPrefix)
			if err != nil {
				return nil, err
			}
			scheduler.RemoveByIdPrefix(jobIDPrefix)
			return nil, nil
		}
}