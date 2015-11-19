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
	hubbleAuth "github.com/Jumpscale/hubble/auth"
	"net/http"
	"crypto/tls"
	"net"
	"github.com/Jumpscale/agentcontroller2/logged"
)

const (
	agentInteractiveAfterOver = 30 * time.Second
)

type Application struct {
	redisPool                *redis.Pool
	internalCommands         *internals.Manager
	commandSource            *logged.CommandSource
	commandResponder         *logged.CommandResponder
	agentCommands            core.AgentCommands
	settings                 *configs.Settings
	scheduler                *scheduling.Scheduler
	rest                     *rest.Manager
	events                   *events.Handler
	executedCommandProcessor commandprocessing.CommandProcessor
	liveAgents               core.AgentInformationStorage
	jumpscriptStore          core.JumpScriptStore

	producers                map[string]chan *core.PollData
	producersLock            sync.Mutex
}


func NewApplication(settingsPath string) *Application {

	settings := loadSettings(settingsPath)
	log.Printf("[+] redis server: <%s>\n", settings.Main.RedisHost)

	redisPool := newRedisPool(settings.Main.RedisHost, settings.Main.RedisPassword)
	panicIfRedisIsNotOK(redisPool)

	app := Application{
		redisPool: redisPool,
		agentCommands: redisdata.AgentCommands(redisPool),
		liveAgents: agentdata.NewAgentData(),
		producers: make(map[string]chan *core.PollData),
		settings: settings,
		jumpscriptStore: redisdata.NewJumpScriptStore(redisPool),
	}

	{
		redisSource := redisdata.NewCommandSource(redisPool)
		interceptedSource := interceptors.NewInterceptedCommandSource(redisSource, app.jumpscriptStore)
		commandLog := redisdata.NewCommandLog(redisPool)
		loggedSource := &logged.CommandSource{
			CommandSource: interceptedSource,
			Log: commandLog,
		}
		app.commandSource = loggedSource
	}

	app.commandResponder = &logged.CommandResponder{
		CommandResponder: redisdata.NewRedisCommandResponder(redisPool),
		Log: redisdata.NewCommandResponseLog(redisPool),
	}

	app.internalCommands = internals.NewManager(app.liveAgents, app.commandResponder)
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
		app.commandResponder,
		app.settings,
		redisdata.NewAgentLog(redisPool),
		app.jumpscriptStore,
	)

	commandProcessor, err := commandprocessing.NewProcessor(
		&app.settings.Processor,
		app.redisPool,
		app.commandSource.Log,
		app.commandResponder.Log,
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

	startHubbleAgent(app.settings.Syncthing.Port)

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
			c, err := redis.DialTimeout("tcp", addr, 0, agentInteractiveAfterOver / 2, 0)

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

func (app *Application) processSingleCommand() {

	command, err := app.commandSource.BlockingPop()
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

	targetAgents := agentsForCommand(app.liveAgents, command)
	if len(targetAgents) == 0 {
		errResponse := errorResponseFor(command, "No matching connected agents found")
		err := app.commandResponder.RespondToCommand(errResponse)
		if err != nil {
			panic("Failed to send error response")
		}
	} else {
		app.distributeCommandToAgents(targetAgents, command)
	}
	app.commandResponder.SignalAsPickedUp(command)
}


func (app *Application) distributeCommandToAgents(agents []core.AgentID, command *core.Command) {

	for _, agentID := range agents {

		log.Println("Dispatching message to", agentID)
		err := app.agentCommands.Enqueue(agentID, command)
		if err != nil {
			log.Println("[-] push error: ", err)
		}

		response := queuedResponseFor(command, agentID)
		app.commandResponder.RespondToCommand(response)
	}
}

func (app *Application) getProducerChan(agentID core.AgentID) chan <- *core.PollData {

	key := fmt.Sprintf("%v:%v", agentID.GID, agentID.NID)

	app.producersLock.Lock()
	producer, ok := app.producers[key]
	if !ok {
		//start routine for this agent.
		log.Printf("Agent %v active, starting agent routine\n", agentID)

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

					app.liveAgents.SetRoles(agentID, data.Roles)

					pendingCommand, err := app.agentCommands.BlockingDequeue(agentID)
					if err != nil {
						if !core.IsTimeout(err) {
							log.Println("Couldn't get new job for agent", key, err)
						}

						return true
					}

					select {
					case msgChan <- string(pendingCommand.JSON):
					//caller consumed this job, it's safe to set it's state to RUNNING now.
						response := runningResponseFor(pendingCommand, agentID)
						app.commandResponder.RespondToCommand(response)
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
