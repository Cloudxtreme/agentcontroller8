package application
import (
	"github.com/Jumpscale/agentcontroller2/core"
	"github.com/Jumpscale/agentcontroller2/internals"
	"github.com/Jumpscale/agentcontroller2/interceptors"
	"github.com/garyburd/redigo/redis"
	"github.com/Jumpscale/agentcontroller2/configs"
	"github.com/Jumpscale/agentcontroller2/scheduling"
	"github.com/Jumpscale/agentcontroller2/events"
	"sync"
	"github.com/Jumpscale/agentcontroller2/rest"
	"github.com/Jumpscale/agentcontroller2/commandprocessing"
	"github.com/Jumpscale/agentcontroller2/redisdata"
	"log"
	"github.com/Jumpscale/agentcontroller2/agentdata"
	"fmt"
	"time"
	"container/list"
	"math/rand"
	"strconv"
	hubbleAuth "github.com/Jumpscale/hubble/auth"
	"net/http"
	"crypto/tls"
	"net"
)

const (
	agentInteractiveAfterOver = 30 * time.Second
)

type Application struct {
	redisPool                *redis.Pool
	internalCommands         *internals.Manager
	commandSource            core.CommandSource
	outgoing                 core.Outgoing
	executedCommands         core.LoggedCommands
	executedCommandsResults  core.LoggedCommandResults
	agentCommands            core.AgentCommands
	settings                 *configs.Settings
	scheduler                *scheduling.Scheduler
	rest                     *rest.Manager
	events                   *events.Handler
	executedCommandProcessor commandprocessing.CommandProcessor
	liveAgents               core.AgentInformationStorage

	producers                map[string]chan *core.PollData
	producersLock            sync.Mutex
}


func NewApplication(settingsPath string) *Application {

	settings := loadSettings(settingsPath)
	log.Printf("[+] redis server: <%s>\n", settings.Main.RedisHost)

	redisPool := newRedisPool(settings.Main.RedisHost, settings.Main.RedisPassword)
	panicIfRedisIsNotOK(redisPool)

	app := Application {
		redisPool: redisPool,
		commandSource: interceptors.Intercept(redisdata.CommandSource(redisPool), redisPool),
		outgoing: redisdata.Outgoing(redisPool),
		executedCommands: redisdata.LoggedCommands(redisPool),
		executedCommandsResults: redisdata.LoggedCommandResult(redisPool),
		agentCommands: redisdata.AgentCommands(redisPool),
		liveAgents: agentdata.NewAgentData(),
		producers: make(map[string]chan* core.PollData),
		settings: settings,
	}

	app.internalCommands = internals.NewManager(app.liveAgents, app.outgoing, app.sendResult)
	app.scheduler = scheduling.NewScheduler(app.redisPool, app.commandSource)

	app.internalCommands.RegisterProcessor("scheduler_add", app.scheduler.Add)
	app.internalCommands.RegisterProcessor("scheduler_list", app.scheduler.List)
	app.internalCommands.RegisterProcessor("scheduler_remove", app.scheduler.Remove)
	app.internalCommands.RegisterProcessor("scheduler_remove_prefix", app.scheduler.RemovePrefix)

	eventHandler, err := events.NewEventsHandler(&app.settings.Events, app.getProducerChan)
	if err != nil {
		log.Fatal("Failed to load events handlers module", err)
	}
	app.events = eventHandler

	app.rest = rest.NewManager(
		app.events,
		app.getProducerChan,
		app.redisPool,
		app.sendResult,
		app.settings,
	)

	commandProcessor, err := commandprocessing.NewProcessor(
		&app.settings.Processor,
		app.redisPool,
		app.executedCommands,
		app.executedCommandsResults,
	)
	if err != nil {
		log.Fatal("Failed to load processors module", err)
	}
	app.executedCommandProcessor = commandProcessor

	return &app
}

func (app *Application) Run() {

	go func() {
		for {
			app.processSingleCommand()
		}
	}()

	app.scheduler.Start()

	app.executedCommandProcessor.Start()

	hubbleAuth.Install(hubbleAuth.NewAcceptAllModule())

	var wg sync.WaitGroup
	wg.Add(len(app.settings.Listen))
	for _, httpBinding := range app.settings.Listen {
		go func(httpBinding configs.HTTPBinding) {
			server := &http.Server{Addr: httpBinding.Address, Handler: app.rest.Engine()}
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

	startSyncthingHubbleAgent(app.settings.Syncthing.Port)

	wg.Wait()
}

func loadSettings(settingsPath string) *configs.Settings {
	settings, err := configs.LoadSettingsFromTomlFile(settingsPath)
	if err != nil {
		log.Fatal("Error loading configuration file:", err)
	}
	return settings
}

func panicIfRedisIsNotOK(redisConnPool *redis.Pool) {
	db := redisConnPool.Get()
	defer db.Close()

	if _, err := db.Do("PING"); err != nil {
		panic(fmt.Sprintf("Failed to connect to redis: %v", err))
	}
}


func newRedisPool(addr string, password string) *redis.Pool {
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

func (app *Application) getActiveAgents(onlyGid int, roles []string) []core.AgentID {
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

	return app.liveAgents.FilteredConnectedAgents(gidFilter, roleFilter)
}




func (app *Application) sendResult(result *core.CommandResult) error {

	// Respond
	err := app.outgoing.RespondToCommand(result)
	if err != nil {
		return err
	}

	// Log for processing
	err = app.executedCommandsResults.Push(result)
	if err != nil {
		return err
	}

	return nil
}

func (app *Application) processSingleCommand() {

	command, err := app.commandSource.Pop()
	if err != nil {
		if core.IsTimeout(err) {
			return
		}

		log.Fatal("Coulnd't read new commands from redis", err)
	}

	log.Println("Received command:", command)

	if command.IsInternal() {
		go app.internalCommands.ExecuteInternalCommand(command)
		return
	}

	targetAgents := app.agentsForCommand(command)
	app.distributeCommandToAgents(targetAgents, command)

	err = app.executedCommands.Push(command)
	if err != nil {
		log.Println("[-] log push error: ", err)
	}

	app.outgoing.SignalAsQueued(command)
}

func (app *Application) agentsForCommand(command *core.Command) []core.AgentID {

	ids := list.New()

	if len(command.Content.Roles) > 0 {
		//command has a given role
		activeAgents := app.getActiveAgents(command.Content.Gid, command.Content.Roles)
		if len(activeAgents) == 0 {
			//no active agents that saticifies this role.
			result := &core.CommandResultContent{
				ID:        command.Content.ID,
				Gid:       command.Content.Gid,
				Nid:       command.Content.Nid,
				Tags:      command.Content.Tags,
				State:     core.CommandStateError,
				Data:      fmt.Sprintf("No agents with role '%v' alive!", command.Content.Roles),
				StartTime: int64(time.Duration(time.Now().UnixNano()) / time.Millisecond),
			}

			resultMessage, err := core.CommandResultFromCommandResultContent(result)
			if err != nil {
				panic(err)
			}

			app.sendResult(resultMessage)
		} else {
			if command.Content.Fanout {
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
		key := fmt.Sprintf("%d:%d", command.Content.Gid, command.Content.Nid)
		_, ok := app.producers[key]
		if !ok {
			//send error message to
			result := &core.CommandResultContent{
				ID:        command.Content.ID,
				Gid:       command.Content.Gid,
				Nid:       command.Content.Nid,
				Tags:      command.Content.Tags,
				State:     core.CommandStateError,
				Data:      fmt.Sprintf("Agent is not alive!"),
				StartTime: int64(time.Duration(time.Now().UnixNano()) / time.Millisecond),
			}

			resultMessage, err := core.CommandResultFromCommandResultContent(result)
			if err != nil {
				panic(err)
			}

			app.sendResult(resultMessage)
		} else {
			ids.PushBack(core.AgentID{GID: uint(command.Content.Gid), NID: uint(command.Content.Nid)})
		}
	}

	var agents []core.AgentID
	for e := ids.Front(); e != nil; e = e.Next() {
		agents = append(agents, e.Value.(core.AgentID))
	}

	return agents
}

func (app *Application) distributeCommandToAgents(agents []core.AgentID, command *core.Command) {

	for _, agentID := range agents {

		resultContent := core.CommandResultContent{
			ID:        command.Content.ID,
			Gid:       int(agentID.GID),
			Nid:       int(agentID.NID),
			Tags:      command.Content.Tags,
			State:     core.CommandStateQueued,
			StartTime: int64(time.Duration(time.Now().UnixNano()) / time.Millisecond),
		}

		result, err := core.CommandResultFromCommandResultContent(&resultContent)
		if err != nil {
			panic(err)
		}

		err = app.outgoing.RespondToCommand(result)
		if err != nil {
			log.Println("[-] failsed to respond with", result)
		}

		log.Println("Dispatching message to", agentID)
		err = app.agentCommands.Enqueue(agentID, command)
		if err != nil {
			log.Println("[-] push error: ", err)
		}

		app.executedCommandsResults.Push(result)
	}
}

func (app *Application) getProducerChan(gid string, nid string) chan<- *core.PollData {
	key := fmt.Sprintf("%s:%s", gid, nid)

	app.producersLock.Lock()
	producer, ok := app.producers[key]
	if !ok {
		igid, _ := strconv.Atoi(gid)
		inid, _ := strconv.Atoi(nid)
		agentID := core.AgentID{GID: uint(igid), NID: uint(inid)}
		//start routine for this agent.
		log.Printf("Agent %s:%s active, starting agent routine\n", gid, nid)

		producer = make(chan *core.PollData)
		app.producers[key] = producer
		go func() {

			defer func() {

				//no agent tried to connect
				close(producer)
				app.producersLock.Lock()
				defer app.producersLock.Unlock()
				delete(app.producers, key)
				app.liveAgents.DropAgent(agentID)
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
					app.liveAgents.SetRoles(agentID, agentRoles)

					pendingCommand, err := app.agentCommands.Dequeue(agentID)
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

						err = app.outgoing.RespondToCommand(result)
						if err != nil {
							log.Println("[-] failed to respond with", result)
						}
						app.executedCommandsResults.Push(result)
					default:
					//caller didn't want to receive this command. have to repush it
					//directly on the agent queue. to avoid doing the redispatching.
						app.agentCommands.ReportUnexecutedCommand(pendingCommand, agentID)
					}

					return true
				}() {
					return
				}
			}

		}()
	}
	app.producersLock.Unlock()

	return producer
}
