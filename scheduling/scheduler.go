package scheduling

import (
	"encoding/json"
	"github.com/Jumpscale/agentcontroller8/core"
	"github.com/Jumpscale/agentcontroller8/redisdata/ds"
	"github.com/garyburd/redigo/redis"
	"github.com/robfig/cron"
	"log"
	"strings"
)

//Scheduler schedules cron jobs
type Scheduler struct {
	cron            *cron.Cron
	pool            *redis.Pool
	commandPipeline core.CommandSource
	commands        ds.Hash
}

func NewScheduler(pool *redis.Pool, commandPipeline core.CommandSource) *Scheduler {
	sched := &Scheduler{
		cron:            cron.New(),
		pool:            pool,
		commandPipeline: commandPipeline,
		commands:        ds.GetHash("controller.schedule"),
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

	//we can safely push the command to the hashset now.
	err = sched.commands.Set(sched.pool, job.ID, JobToJSON(job))
	return err
}

// Lists all scheduled jobs
func (sched *Scheduler) ListJobs() []Job {
	jobsMap, err := sched.commands.ToStringMap(sched.pool)
	if err != nil {
		log.Fatalf("Redis failure: %v", err)
	}

	jobs := make([]Job, 0)
	for _, jsonJob := range jobsMap {
		job, err := JobFromJSON([]byte(jsonJob))
		if err != nil {
			log.Fatal("Corrupted job stored in Redis")
		}
		jobs = append(jobs, *job)
	}

	return jobs
}

func (sched *Scheduler) RemoveByID(id string) (int, error) {
	deleted, err := sched.commands.Delete(sched.pool, id)
	if !deleted {
		return 0, err
	}
	sched.restart()
	return 1, err
}

// Removes all scheduled jobs that have the given ID prefix
func (sched *Scheduler) RemoveByIdPrefix(idPrefix string) {
	db := sched.pool.Get()
	defer db.Close()

	restart := false
	var cursor int
	for {
		slice, err := redis.Values(db.Do("HSCAN", sched.commands.Name, cursor))
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
					sched.commands.Delete(sched.pool, key)
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
		slice, err := redis.Values(db.Do("HSCAN", sched.commands.Name, cursor))
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
