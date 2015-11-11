package main

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/pborman/uuid"
	"github.com/robfig/cron"
	"log"
	"strings"
	"github.com/Jumpscale/agentcontroller2/internals"
	"github.com/Jumpscale/agentcontroller2/messages"
)

const (
	hashScheduleKey = "controller.schedule"
)

//Scheduler schedules cron jobs
type Scheduler struct {
	cron *cron.Cron
	pool *redis.Pool
}

//SchedulerJob represented a shceduled job as stored in redis
type SchedulerJob struct {
	ID   string                 `json:"id"`
	Cron string                 `json:"cron"`
	Cmd  map[string]interface{} `json:"cmd"`
}

//Run runs the scheduled job
func (job *SchedulerJob) Run() {

	job.Cmd["id"] = uuid.New()

	dump, _ := json.Marshal(job.Cmd)

	log.Println("Scheduler: Running job", job.ID, job.Cmd["id"])

	command, err := messages.CommandMessageFromJSON(dump)
	if err != nil {
		panic(err)
	}

	err = incomingCommands.Push(command)
	if err != nil {
		log.Println("Failed to run scheduled command", job.ID)
	}
}

//NewScheduler created a new instance of the scheduler
func NewScheduler(pool *redis.Pool) *Scheduler {
	sched := &Scheduler{
		cron: cron.New(),
		pool: pool,
	}

	return sched
}

//Add create a schdule with the cmd ID (overrides old ones)
func (sched *Scheduler) Add(_ *internals.Manager, cmd *messages.CommandMessage) (interface{}, error) {
	defer sched.restart()

	db := sched.pool.Get()
	defer db.Close()

	job := &SchedulerJob{}

	err := json.Unmarshal([]byte(cmd.Content.Data), job)

	if err != nil {
		log.Println("Failed to load command spec", cmd.Content.Data, err)
		return nil, err
	}

	_, err = cron.Parse(job.Cron)
	if err != nil {
		return nil, err
	}

	job.ID = cmd.Content.ID
	//we can safely push the command to the hashset now.
	db.Do("HSET", hashScheduleKey, cmd.Content.ID, cmd.Content.Data)

	return true, nil
}

//List lists all scheduled jobs
func (sched *Scheduler) List(_ *internals.Manager, _ *messages.CommandMessage) (interface{}, error) {
	db := sched.pool.Get()
	defer db.Close()

	return redis.StringMap(db.Do("HGETALL", hashScheduleKey))
}

//Remove removes the scheduled job that has this cmd.ID
func (sched *Scheduler) Remove(_ *internals.Manager, cmd *messages.CommandMessage) (interface{}, error) {
	db := sched.pool.Get()
	defer db.Close()

	value, err := redis.Int(db.Do("HDEL", hashScheduleKey, cmd.Content.ID))

	if value > 0 {
		//actuall job was deleted. need to restart the scheduler
		sched.restart()
	}

	return value, err
}

//RemovePrefix removes all scheduled jobs that has the cmd.ID as a prefix
func (sched *Scheduler) RemovePrefix(_ *internals.Manager, cmd *messages.CommandMessage) (interface{}, error) {
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
				job := &SchedulerJob{}

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
