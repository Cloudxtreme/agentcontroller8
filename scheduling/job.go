package scheduling

import (
	"encoding/json"
	"github.com/Jumpscale/agentcontroller8/core"
	"github.com/pborman/uuid"
	"log"
)

//SchedulerJob represented a shceduled job as stored in redis
type Job struct {
	ID string `json:"id"`

	// The cron-style spec of the scheduled times
	Cron string `json:"cron"`

	// The RawCommand of the command being executed
	Cmd map[string]interface{} `json:"cmd"`

	// The job will be executed by being pushed to this pipeline
	commandPipeline core.CommandSource
}

func (job *Job) Run() {

	job.Cmd["id"] = uuid.New()

	dump, _ := json.Marshal(job.Cmd)

	log.Println("Scheduler: Running job", job.ID, job.Cmd["id"])

	command, err := core.CommandFromJSON(dump)
	if err != nil {
		log.Fatal(err)
	}

	err = job.commandPipeline.Push(command)
	if err != nil {
		log.Println("Failed to run scheduled command", job.ID)
	}
}

func JobFromJSON(data []byte) (*Job, error) {
	job := Job{}
	err := json.Unmarshal(data, &job)
	if err != nil {
		return nil, err
	}
	return &job, err
}

func JobToJSON(job *Job) []byte {
	jsonData, err := json.Marshal(job)
	if err != nil {
		log.Fatalf("Failed to JSON Marshal: %v", err)
	}
	return jsonData
}
