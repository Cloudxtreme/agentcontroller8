package main

import (
	"container/list"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Jumpscale/agentcontroller2/agentdata"
	"github.com/Jumpscale/agentcontroller2/configs"
	"github.com/Jumpscale/agentcontroller2/core"
	"github.com/Jumpscale/agentcontroller2/events"
	"github.com/Jumpscale/agentcontroller2/processors"
	"github.com/Jumpscale/agentcontroller2/rest"
	hublleAgent "github.com/Jumpscale/hubble/agent"
	hubbleAuth "github.com/Jumpscale/hubble/auth"
	"github.com/garyburd/redigo/redis"
)

const (
	agentInteractiveAfterOver = 30 * time.Second
	roleAll                   = "*"
	cmdQueueMain              = "cmds.queue"
	resultsQueueMain          = "resutls.queue"
	cmdQueueCmdQueued         = "cmd.%s.queued"
	cmdQueueAgentResponse     = "cmd.%s.%d.%d"
	logQueue                  = "joblog"
	hashCmdResults            = "jobresult:%s"
	cmdInternal               = "controller"
)

// redis stuff
func newPool(addr string, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialTimeout("tcp", addr, 0, agentInteractiveAfterOver/2, 0)

			if err != nil {
				panic(err.Error())
			}

			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}

			return c, err
		},
	}
}

var pool *redis.Pool

func getAgentQueue(id core.AgentID) string {
	return fmt.Sprintf("cmds:%d:%d", id.GID, id.NID)
}

func getAgentResultQueue(result *core.CommandResult) string {
	return fmt.Sprintf(cmdQueueAgentResponse, result.ID, result.Gid, result.Nid)
}

func getActiveAgents(onlyGid int, roles []string) [][]int {
	var gidFilter *uint
	var roleFilter []core.AgentRole

	if onlyGid > 0 {
		filterValue := uint(onlyGid)
		gidFilter = &filterValue
	}

	if len(roles) > 0 {
		for _, roleStr := range roles {
			roleFilter = append(roleFilter, core.AgentRole(roleStr))
		}
	}

	connectedAgents := liveAgents.FilteredConnectedAgents(gidFilter, roleFilter)

	var output [][]int
	for _, connectedAgent := range connectedAgents {
		output = append(output, []int{int(connectedAgent.GID), int(connectedAgent.NID)})
	}

	return output
}

func sendResult(result *core.CommandResult) error {
	db := pool.Get()
	defer db.Close()

	key := fmt.Sprintf("%d:%d", result.Gid, result.Nid)
	if data, err := json.Marshal(&result); err == nil {
		err = db.Send("HSET",
			fmt.Sprintf(hashCmdResults, result.ID),
			key,
			data)

		if err != nil {
			return err
		}
		// push message to client result queue queue
		err = db.Send("RPUSH", getAgentResultQueue(result), data)
		if err != nil {
			return err
		}

		//main results queue for results processors
		err = db.Send("RPUSH", resultsQueueMain, data)
		if err != nil {
			return err
		}
	} else {
		return err
	}

	return nil
}

// Caller is expecting a map with keys "GID:NID" of each live agent and values being
// the sequence of roles the agent declares.
func internalListAgents(cmd *core.Command) (interface{}, error) {
	output := make(map[string][]string)
	for _, agentID := range liveAgents.ConnectedAgents() {
		var roles []string
		for _, role := range liveAgents.GetRoles(agentID) {
			roles = append(roles, string(role))
		}
		output[fmt.Sprintf("%d:%d", agentID.GID, agentID.NID)] = roles
	}
	return output, nil
}

var internals = map[string]func(*core.Command) (interface{}, error){
	"list_agents": internalListAgents,
}

func processInternalCommand(command core.Command) {
	result := &core.CommandResult{
		ID:        command.ID,
		Gid:       command.Gid,
		Nid:       command.Nid,
		Tags:      command.Tags,
		State:     core.CommandStateError,
		StartTime: int64(time.Duration(time.Now().UnixNano()) / time.Millisecond),
	}

	internal, ok := internals[command.Args.Name]
	if ok {
		data, err := internal(&command)
		if err != nil {
			result.Data = err.Error()
		} else {
			serialized, err := json.Marshal(data)
			if err != nil {
				result.Data = err.Error()
			}
			result.State = "SUCCESS"
			result.Data = string(serialized)
			result.Level = 20
		}
	} else {
		result.State = "UNKNOWN_CMD"
	}

	sendResult(result)
	signalQueues(command.ID)
}

func signalQueues(id string) {
	db := pool.Get()
	defer db.Close()
	db.Do("RPUSH", fmt.Sprintf(cmdQueueCmdQueued, id), "queued")
}

func readSingleCmd() bool {
	db := pool.Get()
	defer db.Close()

	commandEntry, err := redis.Strings(db.Do("BLPOP", cmdQueueMain, "0"))

	if err != nil {
		if core.IsTimeout(err) {
			return true
		}

		log.Fatal("Coulnd't read new commands from redis", err)
	}
	command := commandEntry[1]

	log.Println("Received message:", command)
	command = InterceptCommand(command)
	// parsing json data
	var payload core.Command
	err = json.Unmarshal([]byte(command), &payload)

	if err != nil {
		log.Println("message decoding error:", err)
		return true
	}

	if payload.Cmd == cmdInternal {
		go processInternalCommand(payload)
		return true
	}

	//sort command to the consumer queue.
	//either by role or by the gid/nid.
	ids := list.New()

	if len(payload.Roles) > 0 {
		//command has a given role
		active := getActiveAgents(payload.Gid, payload.Roles)
		if len(active) == 0 {
			//no active agents that saticifies this role.
			result := &core.CommandResult{
				ID:        payload.ID,
				Gid:       payload.Gid,
				Nid:       payload.Nid,
				Tags:      payload.Tags,
				State:     "ERROR",
				Data:      fmt.Sprintf("No agents with role '%v' alive!", payload.Roles),
				StartTime: int64(time.Duration(time.Now().UnixNano()) / time.Millisecond),
			}

			sendResult(result)
		} else {
			if payload.Fanout {
				//fanning out.
				for _, agent := range active {
					ids.PushBack(agent)
				}

			} else {
				agent := active[rand.Intn(len(active))]
				ids.PushBack(agent)
			}
		}
	} else {
		key := fmt.Sprintf("%d:%d", payload.Gid, payload.Nid)
		_, ok := producers[key]
		if !ok {
			//send error message to
			result := &core.CommandResult{
				ID:        payload.ID,
				Gid:       payload.Gid,
				Nid:       payload.Nid,
				Tags:      payload.Tags,
				State:     "ERROR",
				Data:      fmt.Sprintf("Agent is not alive!"),
				StartTime: int64(time.Duration(time.Now().UnixNano()) / time.Millisecond),
			}

			sendResult(result)
		} else {
			ids.PushBack([]int{payload.Gid, payload.Nid})
		}
	}

	// push logs
	if _, err := db.Do("LPUSH", logQueue, command); err != nil {
		log.Println("[-] log push error: ", err)
	}

	//distribution to agents.
	for e := ids.Front(); e != nil; e = e.Next() {
		// push message to client queue
		agent := e.Value.([]int)
		gid := agent[0]
		nid := agent[1]
		agentID := core.AgentID{GID: uint(gid), NID: uint(nid)}

		log.Println("Dispatching message to", agent)
		if _, err := db.Do("RPUSH", getAgentQueue(agentID), command); err != nil {
			log.Println("[-] push error: ", err)
		}

		resultPlaceholder := core.CommandResult{
			ID:        payload.ID,
			Gid:       gid,
			Nid:       nid,
			Tags:      payload.Tags,
			State:     core.CommandStateQueued,
			StartTime: int64(time.Duration(time.Now().UnixNano()) / time.Millisecond),
		}

		if data, err := json.Marshal(&resultPlaceholder); err == nil {
			db.Do("HSET",
				fmt.Sprintf(hashCmdResults, payload.ID),
				fmt.Sprintf("%d:%d", gid, nid),
				data)
		}
	}

	signalQueues(payload.ID)
	return true
}

// Command Reader
func cmdreader() {
	for {
		// waiting message from master queue
		if !readSingleCmd() {
			return
		}
	}
}

var producers = make(map[string]chan *core.PollData)
var liveAgents = agentdata.NewAgentData()

// var activeRoles map[string]int = make(map[string]int)
// var activeGridRoles map[string]map[string]int = make(map[string]map[string]int)
var producersLock sync.Mutex

func getProducerChan(gid string, nid string) chan<- *core.PollData {
	key := fmt.Sprintf("%s:%s", gid, nid)

	producersLock.Lock()
	producer, ok := producers[key]
	if !ok {
		igid, _ := strconv.Atoi(gid)
		inid, _ := strconv.Atoi(nid)
		agentID := core.AgentID{GID: uint(igid), NID: uint(inid)}
		//start routine for this agent.
		log.Printf("Agent %s:%s active, starting agent routine\n", gid, nid)

		producer = make(chan *core.PollData)
		producers[key] = producer
		go func() {
			//db := pool.Get()

			defer func() {
				//db.Close()

				//no agent tried to connect
				close(producer)
				producersLock.Lock()
				defer producersLock.Unlock()
				delete(producers, key)
				liveAgents.DropAgent(agentID)
			}()

			for {
				if !func() bool {
					var data *core.PollData

					select {
					case data = <-producer:
					case <-time.After(agentInteractiveAfterOver):
						//no active agent for 10 min
						log.Println("Agent", key, "is inactive for over ", agentInteractiveAfterOver, ", cleaning up.")
						return false
					}

					msgChan := data.MsgChan
					defer close(msgChan)

					roles := data.Roles

					var agentRoles []core.AgentRole
					for _, role := range roles {
						agentRoles = append(agentRoles, core.AgentRole(role))
					}
					liveAgents.SetRoles(agentID, agentRoles)

					db := pool.Get()
					defer db.Close()

					pending, err := redis.Strings(db.Do("BLPOP", getAgentQueue(agentID), "0"))
					if err != nil {
						if !core.IsTimeout(err) {
							log.Println("Couldn't get new job for agent", key, err)
						}

						return true
					}

					select {
					case msgChan <- pending[1]:
						//caller consumed this job, it's safe to set it's state to RUNNING now.
						var payload core.Command
						if err := json.Unmarshal([]byte(pending[1]), &payload); err != nil {
							break
						}

						resultPlacehoder := core.CommandResult{
							ID:        payload.ID,
							Gid:       igid,
							Nid:       inid,
							Tags:      payload.Tags,
							State:     core.CommandStateRunning,
							StartTime: int64(time.Duration(time.Now().UnixNano()) / time.Millisecond),
						}

						if data, err := json.Marshal(&resultPlacehoder); err == nil {
							db.Do("HSET",
								fmt.Sprintf(hashCmdResults, payload.ID),
								key,
								data)
						}
					default:
						//caller didn't want to receive this command. have to repush it
						//directly on the agent queue. to avoid doing the redispatching.
						if pending[1] != "" {
							db.Do("LPUSH", getAgentQueue(agentID), pending[1])
						}
					}

					return true
				}() {
					return
				}
			}

		}()
	}
	producersLock.Unlock()

	return producer
}

//StartSyncthingHubbleAgent start the builtin hubble agent required for Syncthing
func StartSyncthingHubbleAgent(hubblePort int) {

	wsURL := fmt.Sprintf("ws://127.0.0.1:%d/0/0/hubble", hubblePort)
	log.Println("Starting local hubble agent at", wsURL)
	agent := hublleAgent.NewAgent(wsURL, "controller", "", nil)
	var onExit func(agt hublleAgent.Agent, err error)

	onExit = func(agt hublleAgent.Agent, err error) {
		if err != nil {
			go func() {
				time.Sleep(3 * time.Second)
				agt.Start(onExit)
			}()
		}
	}

	agent.Start(onExit)
}

var settings configs.Settings

func main() {
	var cfg string

	flag.StringVar(&cfg, "c", "", "Path to config file")
	flag.Parse()

	if cfg == "" {
		log.Println("Missing required option -c")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var err error
	settings, err = configs.LoadSettingsFromTomlFile(cfg)
	if err != nil {
		log.Panicln("Error loading configuration file:", err)
	}

	log.Printf("[+] redis server: <%s>\n", settings.Main.RedisHost)

	pool = newPool(settings.Main.RedisHost, settings.Main.RedisPassword)

	db := pool.Get()
	if _, err := db.Do("PING"); err != nil {
		panic(fmt.Sprintf("Failed to connect to redis: %v", err))
	}

	db.Close()

	go cmdreader()

	//start schedular.
	scheduler := NewScheduler(pool)
	internals["scheduler_add"] = scheduler.Add
	internals["scheduler_list"] = scheduler.List
	internals["scheduler_remove"] = scheduler.Remove
	internals["scheduler_remove_prefix"] = scheduler.RemovePrefix

	scheduler.Start()

	eventHandler, err := events.NewEventsHandler(&settings.Events, getProducerChan)
	if err != nil {
		log.Fatal("Failed to load events handlers module", err)
	}

	restInterface := rest.NewManager(
		eventHandler,
		getProducerChan,
		pool,
		sendResult,
		&settings,
	)

	//start results processors
	processor, err := processors.NewResultsProcessor(&settings.Processors, pool, resultsQueueMain)
	if err != nil {
		log.Fatal("Failed to load processors module", err)
	}

	processor.Start()

	hubbleAuth.Install(hubbleAuth.NewAcceptAllModule())

	var wg sync.WaitGroup
	wg.Add(len(settings.Listen))
	for _, httpBinding := range settings.Listen {
		go func(httpBinding configs.HTTPBinding) {
			server := &http.Server{Addr: httpBinding.Address, Handler: restInterface.Engine()}
			if httpBinding.TLSEnabled() {
				server.TLSConfig = &tls.Config{}

				if err := configureServerCertificates(httpBinding, server); err != nil {
					log.Panicln("Unable to load the server certificates", err)
				}

				if err := configureClientCertificates(httpBinding, server); err != nil {
					log.Panicln("Unable to load the clientCA's", err)
				}

				ln, err := net.Listen("tcp", server.Addr)
				if err != nil {
					log.Panicln(err)
				}

				tlsListener := tls.NewListener(ln, server.TLSConfig)
				log.Println("Listening on", httpBinding.Address, "with TLS")
				if err := server.Serve(tlsListener); err != nil {
					log.Panicln(err)
				}
				wg.Done()
			} else {
				log.Println("Listening on", httpBinding.Address)
				if err := server.ListenAndServe(); err != nil {
					log.Panicln(err)
				}
				wg.Done()
			}
		}(httpBinding)
	}

	StartSyncthingHubbleAgent(settings.Syncthing.Port)

	wg.Wait()
}
