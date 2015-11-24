package scheduling

import (
	"encoding/json"
	"github.com/Jumpscale/agentcontroller2/core"
	"github.com/garyburd/redigo/redis"
	"github.com/robfig/cron"
	"log"
	"strings"
	"fmt"
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

func NewScheduler(pool *redis.Pool, commandPipeline core.CommandSource) *Scheduler {
	sched := &Scheduler{
		cron:            cron.New(),
		pool:            pool,
		commandPipeline: commandPipeline,
	}

	return sched
}

// Returns an error on an invalid Cron timing spec
func validateCronSpec(timing string) error {
	_, err := cron.Parse(timing)
	if err != nil {
		return err
	}
	return nil
}

// Adds a job to the scheduler (overrides old ones)
func (sched *Scheduler) AddJob(job *Job) error {

	err := validateCronSpec(job.Cron)
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


// Lists all scheduled jobs
func (sched *Scheduler) ListJobs() []Job {
	db := sched.pool.Get()
	defer db.Close()

	jobsMap, err := redis.StringMap(db.Do("HGETALL", hashScheduleKey))
	if err != nil {
		panic(fmt.Errorf("Redis failure: %v", err))
	}

	var jobs []Job
	for _, jsonJob := range jobsMap {
		job, err := JobFromJSON([]byte(jsonJob))
		if err != nil {
			panic("Corrupted job stored in Redis")
		}
		jobs = append(jobs, *job)
	}

	return jobs
}


func (sched *Scheduler) RemoveByID(id string) (int, error) {
	db := sched.pool.Get()
	defer db.Close()

	value, err := redis.Int(db.Do("HDEL", hashScheduleKey, id))

	if value > 0 {
		//actuall job was deleted. need to restart the scheduler
		sched.restart()
	}

	return value, err
}

// Removes all scheduled jobs that have the given ID prefix
func (sched *Scheduler) RemoveByIdPrefix(idPrefix string) {
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
				if strings.Index(key, idPrefix) == 0 {
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
