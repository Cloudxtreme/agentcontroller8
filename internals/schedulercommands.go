package internals
import (
	"github.com/Jumpscale/agentcontroller2/scheduling"
	"github.com/Jumpscale/agentcontroller2/core"
)

func (manager *Manager) setUpSchedulerCommands(scheduler *scheduling.Scheduler) {

	manager.commandHandlers[SchedulerAddJob] =
		func(cmd *core.Command) (interface{}, error) {
			job, err := scheduling.JobFromJSON([]byte(cmd.Content.Data))
			if err != nil {
				return nil, err
			}
			job.ID = cmd.Content.ID		// Essential, this is how the Job gets its ID
			return nil, scheduler.AddJob(job)
		}

	manager.commandHandlers[SchedulerListJobs] =
		func(_ *core.Command) (interface{}, error) {
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