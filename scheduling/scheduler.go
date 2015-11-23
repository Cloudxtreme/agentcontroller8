package scheduling

import (
	"encoding/json"
	"github.com/Jumpscale/agentcontroller2/core"
	"github.com/garyburd/redigo/redis"
	"github.com/pborman/uuid"
	"github.com/robfig/cron"
	"log"
	"strings"
)

const (
	hashScheduleKey = "controller.schedule"
)

//Scheduler schedules cron jobs
type Scheduler struct {
	cron            *cron.Cron
	pool            *redis.Pool
	commandPipeline core.CommandSource
}

//SchedulerJob represented a shceduled job as stored in redis
type Job struct {
	ID              string                 `json:"id"`
	Cron            string                 `json:"cron"`
	Cmd             map[string]interface{} `json:"cmd"`
	commandPipeline core.CommandSource
}

//Run runs the scheduled job
func (job *Job) Run() {

	job.Cmd["id"] = uuid.New()

	dump, _ := json.Marshal(job.Cmd)

	log.Println("Scheduler: Running job", job.ID, job.Cmd["id"])

	command, err := core.CommandFromJSON(dump)
	if err != nil {
		panic(err)
	}

	err = job.commandPipeline.Push(command)
	if err != nil {
		log.Println("Failed to run scheduled command", job.ID)
	}
}

//NewScheduler created a new instance of the scheduler
func NewScheduler(pool *redis.Pool, commandPipeline core.CommandSource) *Scheduler {
	sched := &Scheduler{
		cron:            cron.New(),
		pool:            pool,
		commandPipeline: commandPipeline,
	}

	return sched
}

//AddJob adds a job to the scheduler (overrides old ones)
func (sched *Scheduler) AddJob(job *Job) error {
	_, err := cron.Parse(job.Cron)
	if err != nil {
		return err
	}

	defer sched.restart()

	db := sched.pool.Get()
	defer db.Close()

	data, err := json.Marshal(job)
	if err != nil {
		return err
	}

	//we can safely push the command to the hashset now.
	_, err = db.Do("HSET", hashScheduleKey, job.ID, data)
	return err
}


//Add create a schdule with the cmd ID (overrides old ones).
//This add method is compatible withe the 'internals' manager interface so it can be
//called remotely via the client.
func (sched *Scheduler) Add(cmd *core.Command) (interface{}, error) {
	job := &Job{}

	err := json.Unmarshal([]byte(cmd.Content.Data), job)

	if err != nil {
		log.Println("Failed to load command spec", cmd.Content.Data, err)
		return nil, err
	}

	job.ID = cmd.Content.ID

	err = sched.AddJob(job)
	if err != nil {
		return nil, err
	}

	return true, nil
}

//List lists all scheduled jobs
func (sched *Scheduler) List() (interface{}, error) {
	db := sched.pool.Get()
	defer db.Close()

	return redis.StringMap(db.Do("HGETALL", hashScheduleKey))
}


func (sched *Scheduler) RemoveID(ID string) (int, error) {
	db := sched.pool.Get()
	defer db.Close()

	value, err := redis.Int(db.Do("HDEL", hashScheduleKey, ID))

	if value > 0 {
		//actuall job was deleted. need to restart the scheduler
		sched.restart()
	}

	return value, err
}

//Remove removes the scheduled job that has this cmd.ID
func (sched *Scheduler) Remove(cmd *core.Command) (interface{}, error) {
	return sched.RemoveID(cmd.Content.ID)
}

//RemovePrefix removes all scheduled jobs that has the cmd.ID as a prefix
func (sched *Scheduler) RemovePrefix(cmd *core.Command) (interface{}, error) {
	db := sched.pool.Get()
	defer db.Close()

	restart := false
	var cursor int
	for {
		slice, err := redis.Values(db.Do("HSCAN", hashScheduleKey, cursor))
		if err != nil {
			log.Println("Failed to load schedule from redis", err)
			break
		}

		var fields interface{}
		if _, err := redis.Scan(slice, &cursor, &fields); err == nil {
			set, _ := redis.StringMap(fields, nil)

			for key := range set {
				log.Println("Deleting cron job:", key)
				if strings.Index(key, cmd.Content.ID) == 0 {
					restart = true
					db.Do("HDEL", hashScheduleKey, key)
				}
			}
		} else {
			log.Println(err)
			break
		}

		if cursor == 0 {
			break
		}
	}

	if restart {
		sched.restart()
	}

	return nil, nil
}

func (sched *Scheduler) restart() {
	sched.cron.Stop()
	sched.cron = cron.New()
	sched.Start()
}

//Start starts the scheduler
func (sched *Scheduler) Start() {
	db := sched.pool.Get()
	defer db.Close()

	var cursor int
	for {
		slice, err := redis.Values(db.Do("HSCAN", hashScheduleKey, cursor))
		if err != nil {
			log.Println("Failed to load schedule from redis", err)
			break
		}

		var fields interface{}
		if _, err := redis.Scan(slice, &cursor, &fields); err == nil {
			set, _ := redis.StringMap(fields, nil)

			for key, cmd := range set {
				job := &Job{
					commandPipeline: sched.commandPipeline,
				}

				err := json.Unmarshal([]byte(cmd), job)

				if err != nil {
					log.Println("Failed to load scheduled job", key, err)
					continue
				}

				job.ID = key
				sched.cron.AddJob(job.Cron, job)
			}
		} else {
			log.Println(err)
			break
		}

		if cursor == 0 {
			break
		}
	}

	sched.cron.Start()
}
