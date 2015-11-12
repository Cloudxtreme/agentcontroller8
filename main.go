package main

import (
	"container/list"
	"crypto/tls"
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
	"github.com/Jumpscale/agentcontroller2/interceptors"
	"github.com/Jumpscale/agentcontroller2/processors"
	"github.com/Jumpscale/agentcontroller2/rest"
	hublleAgent "github.com/Jumpscale/hubble/agent"
	hubbleAuth "github.com/Jumpscale/hubble/auth"
	"github.com/garyburd/redigo/redis"
	"github.com/Jumpscale/agentcontroller2/internals"
	"github.com/Jumpscale/agentcontroller2/redisdata"
	"github.com/Jumpscale/agentcontroller2/scheduling"
)

const (
	agentInteractiveAfterOver = 30 * time.Second
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
var commandInterceptors *interceptors.Manager
var internalCommands *internals.Manager
var incomingCommands core.IncomingCommands
var outgoing core.Outgoing
var loggedCommands core.LoggedCommands
var loggedCommandResults core.LoggedCommandResults
var agentCommands core.AgentCommands


func getActiveAgents(onlyGid int, roles []string) []core.AgentID {
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

	return liveAgents.FilteredConnectedAgents(gidFilter, roleFilter)
}

func sendResult(result *core.CommandResult) error {

	// Respond
	err := outgoing.RespondToCommand(result)
	if err != nil {
		return err
	}

	// Log for processing
	err = loggedCommandResults.Push(result)
	if err != nil {
		return err
	}

	return nil
}

func readSingleCmd() bool {

	command, err := incomingCommands.Pop()
	if err != nil {
		if core.IsTimeout(err) {
			return true
		}

		log.Fatal("Coulnd't read new commands from redis", err)
	}

	log.Println("Received message:", command)

	command = commandInterceptors.Intercept(command)
	var content = command.Content

	if content.Cmd == "controller" {
		go internalCommands.ProcessInternalCommand(command)
		return true
	}

	//sort command to the consumer queue.
	//either by role or by the gid/nid.
	ids := list.New()

	if len(content.Roles) > 0 {
		//command has a given role
		activeAgents := getActiveAgents(content.Gid, content.Roles)
		if len(activeAgents) == 0 {
			//no active agents that saticifies this role.
			result := &core.CommandResultContent{
				ID:        content.ID,
				Gid:       content.Gid,
				Nid:       content.Nid,
				Tags:      content.Tags,
				State:     core.CommandStateError,
				Data:      fmt.Sprintf("No agents with role '%v' alive!", content.Roles),
				StartTime: int64(time.Duration(time.Now().UnixNano()) / time.Millisecond),
			}

			resultMessage, err := core.CommandResultFromCommandResultContent(result)
			if err != nil {
				panic(err)
			}

			sendResult(resultMessage)
		} else {
			if content.Fanout {
				//fanning out.
				for _, agentID := range activeAgents {
					ids.PushBack(agentID)
				}

			} else {
				randomAgent := activeAgents[rand.Intn(len(activeAgents))]
				ids.PushBack(randomAgent)
			}
		}
	} else {
		key := fmt.Sprintf("%d:%d", content.Gid, content.Nid)
		_, ok := producers[key]
		if !ok {
			//send error message to
			result := &core.CommandResultContent{
				ID:        content.ID,
				Gid:       content.Gid,
				Nid:       content.Nid,
				Tags:      content.Tags,
				State:     core.CommandStateError,
				Data:      fmt.Sprintf("Agent is not alive!"),
				StartTime: int64(time.Duration(time.Now().UnixNano()) / time.Millisecond),
			}

			resultMessage, err := core.CommandResultFromCommandResultContent(result)
			if err != nil {
				panic(err)
			}

			sendResult(resultMessage)
		} else {
			ids.PushBack(core.AgentID{GID: uint(content.Gid), NID: uint(content.Nid)})
		}
	}

	// push logs
	err = loggedCommands.Push(command)
	if err != nil {
		log.Println("[-] log push error: ", err)
	}

	//distribution to agents.
	for e := ids.Front(); e != nil; e = e.Next() {
		// push message to client queue
		agentID := e.Value.(core.AgentID)

		resultContent := core.CommandResultContent{
			ID:        content.ID,
			Gid:       int(agentID.GID),
			Nid:       int(agentID.NID),
			Tags:      content.Tags,
			State:     core.CommandStateQueued,
			StartTime: int64(time.Duration(time.Now().UnixNano()) / time.Millisecond),
		}

		result, err := core.CommandResultFromCommandResultContent(&resultContent)
		if err != nil {
			panic(err)
		}

		err = outgoing.RespondToCommand(result)
		if err != nil {
			log.Println("[-] failsed to respond with", result)
		}

		loggedCommandResults.Push(result)

		log.Println("Dispatching message to", agentID)
		err = agentCommands.Enqueue(agentID, command)
		if err != nil {
			log.Println("[-] push error: ", err)
		}
	}

	outgoing.SignalAsQueued(command)
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

			defer func() {

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

					pendingCommand, err := agentCommands.Dequeue(agentID)
					if err != nil {
						if !core.IsTimeout(err) {
							log.Println("Couldn't get new job for agent", key, err)
						}

						return true
					}

					select {
					case msgChan <- string(pendingCommand.JSON):

						//caller consumed this job, it's safe to set it's state to RUNNING now.

						resultContent := core.CommandResultContent{
							ID:        pendingCommand.Content.ID,
							Gid:       igid,
							Nid:       inid,
							Tags:      pendingCommand.Content.Tags,
							State:     core.CommandStateRunning,
							StartTime: int64(time.Duration(time.Now().UnixNano()) / time.Millisecond),
						}

						result, err :=
							core.CommandResultFromCommandResultContent(&resultContent)
						if err != nil {
							panic(err)
						}

						err = outgoing.RespondToCommand(result)
						if err != nil {
							log.Println("[-] failed to respond with", result)
						}
						loggedCommandResults.Push(result)
					default:
						//caller didn't want to receive this command. have to repush it
						//directly on the agent queue. to avoid doing the redispatching.
						agentCommands.ReportUnexecutedCommand(pendingCommand, agentID)
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
	commandInterceptors = interceptors.NewManager(pool)
	internalCommands = internals.NewManager(liveAgents, outgoing, sendResult)
	incomingCommands = redisdata.IncomingCommands(pool)
	outgoing = redisdata.Outgoing(pool)
	loggedCommands = redisdata.LoggedCommands(pool)
	loggedCommandResults = redisdata.LoggedCommandResult(pool)
	agentCommands = redisdata.AgentCommands(pool)

	db := pool.Get()
	if _, err := db.Do("PING"); err != nil {
		panic(fmt.Sprintf("Failed to connect to redis: %v", err))
	}

	db.Close()

	go cmdreader()

	//start schedular.
	scheduler := scheduling.NewScheduler(pool, incomingCommands)
	internalCommands.RegisterProcessor("scheduler_add", scheduler.Add)
	internalCommands.RegisterProcessor("scheduler_list", scheduler.List)
	internalCommands.RegisterProcessor("scheduler_remove", scheduler.Remove)
	internalCommands.RegisterProcessor("scheduler_remove_prefix", scheduler.RemovePrefix)

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

	//start external command processors
	processor, err := processors.NewProcessor(&settings.Processor, pool, loggedCommands, loggedCommandResults)
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
