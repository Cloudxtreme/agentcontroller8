package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Jumpscale/agentcontroller2/influxdb-client-0.8.8"
	hubbleAuth "github.com/Jumpscale/hubble/auth"
	hubble "github.com/Jumpscale/hubble/proxy"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"github.com/naoina/toml"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"
)

type Settings struct {
	Main struct {
		Listen        string
		RedisHost     string
		RedisPassword string
	}

	Influxdb struct {
		Host     string
		Db       string
		User     string
		Password string
	}

	Handlers struct {
		Binary string
		Cwd    string
		Env    map[string]string
	}
}

//LoadTomlFile loads toml using "github.com/naoina/toml"
func LoadTomlFile(filename string, v interface{}) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	if err := toml.Unmarshal(buf, v); err != nil {
		panic(err)
	}
}

// data types
type CommandMessage struct {
	Id   string `json:"id"`
	Gid  int    `json:"gid"`
	Nid  int    `json:"nid"`
	Role string `json:"role"`
}

type StatsRequest struct {
	Timestamp int64           `json:"timestamp"`
	Series    [][]interface{} `json:"series"`
}

type EvenRequest struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

// redis stuff
func newPool(addr string, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr)

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

// Command Reader
func cmdreader() {
	db := pool.Get()
	defer db.Close()

	for {
		// waiting message from master queue
		command, err := redis.Strings(db.Do("BLPOP", "cmds_queue", "0"))

		log.Println("[+] message from master redis queue")

		if err != nil {
			log.Println("[-] pop error: ", err)
			continue
		}

		log.Println("[+] message payload: ", command[1])

		// parsing json data
		var payload CommandMessage
		err = json.Unmarshal([]byte(command[1]), &payload)

		if err != nil {
			log.Println("[-] message decoding: ", err)
			continue
		}

		//sort command to the consumer queue.
		//either by role or by the gid/nid.
		var id string
		if payload.Role != "" {
			//command has a given role
			if payload.Gid == 0 {
				id = fmt.Sprintf("cmds_queue_%s", payload.Role)
			} else {
				id = fmt.Sprintf("cmds_queue_%d_%s", payload.Gid, payload.Role)
			}
		} else {
			id = fmt.Sprintf("%d:%d", payload.Gid, payload.Nid)
		}

		log.Printf("[+] message destination [%s]\n", id)

		// push message to client queue
		_, err = db.Do("RPUSH", id, command[1])

		if err != nil {
			log.Println("[-] push error: ", err)
		}
	}
}

var producers map[string]chan *PollData = make(map[string]chan *PollData)
var producersLock sync.Mutex

/**
Gets a chain for the caller to wait on, we return a chan chan string instead
of chan string directly to make sure of the following:
1- The redis pop loop will not try to pop jobs out of the queue until there is a caller waiting
   for new commands
2- Prevent multiple clients polling on a single gid:nid at the same time.
*/
type PollData struct {
	Roles   []string
	MsgChan chan string
}

func getProducerChan(gid string, nid string) chan<- *PollData {
	key := fmt.Sprintf("%s:%s", gid, nid)

	producersLock.Lock()
	producer, ok := producers[key]
	if !ok {
		//start routine for this agent.
		producer = make(chan *PollData)
		producers[key] = producer
		go func() {
			db := pool.Get()
			defer func() {
				db.Close()

				//no agent tried to connect
				close(producer)
				producersLock.Lock()
				defer producersLock.Unlock()
				delete(producers, key)
			}()

			for {
				var data *PollData
				select {
				case data = <-producer:
				case time.After(10 * time.Minute):
					//no active agent for 10 min
					log.Printf("Agent %s:%s is inactive for 10min, cleaning up.\n", gid, nid)
					return
				}

				msgChan := data.MsgChan
				roles := data.Roles
				roles_keys := make([]interface{}, 1, len(roles)*2+2)
				roles_keys[0] = key

				for _, role := range roles {
					roles_keys = append(roles_keys,
						fmt.Sprintf("cmds_queue_%s", role),
						fmt.Sprintf("cmds_queue_%s_%s", gid, role))
				}

				roles_keys = append(roles_keys, "0")
				pending, err := redis.Strings(db.Do("BLPOP", roles_keys...))
				if err != nil {
					log.Println(err)
					return
				}

				select {
				case msgChan <- pending[1]:
				default:
					//caller didn't want to receive this command. have to repush it
					db.Do("LPUSH", "cmds_queue", pending[1])
				}

				close(data.MsgChan)
			}

		}()
	}
	producersLock.Unlock()

	return producer
}

// REST stuff
func cmd(c *gin.Context) {
	gid := c.Param("gid")
	nid := c.Param("nid")

	query := c.Request.URL.Query()
	roles := query["role"]
	log.Printf("[+] gin: execute (gid: %s, nid: %s)\n", gid, nid)

	// listen for http closing
	notify := c.Writer.(http.CloseNotifier).CloseNotify()

	timeout := 60 * time.Second

	producer := getProducerChan(gid, nid)

	data := &PollData{
		Roles:   roles,
		MsgChan: make(chan string),
	}

	select {
	case producer <- data:
	case <-time.After(timeout):
		c.String(http.StatusOK, "")
		return
	}
	//at this point we are sure this is the ONLY agent polling on /gid/nid/cmd

	var payload string

	select {
	case payload = <-data.MsgChan:
	case <-notify:
	case <-time.After(timeout):
	}

	c.String(http.StatusOK, payload)
}

func logs(c *gin.Context) {
	gid := c.Param("gid")
	nid := c.Param("nid")

	db := pool.Get()
	defer db.Close()

	log.Printf("[+] gin: log (gid: %s, nid: %s)\n", gid, nid)

	// read body
	content, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		log.Println("[-] cannot read body:", err)
		c.JSON(http.StatusInternalServerError, "error")
		return
	}

	// push body to redis
	id := fmt.Sprintf("%s:%s:log", gid, nid)
	log.Printf("[+] message destination [%s]\n", id)

	// push message to client queue
	_, err = db.Do("RPUSH", id, content)

	c.JSON(http.StatusOK, "ok")
}

func result(c *gin.Context) {
	gid := c.Param("gid")
	nid := c.Param("nid")

	db := pool.Get()
	defer db.Close()

	log.Printf("[+] gin: result (gid: %s, nid: %s)\n", gid, nid)

	// read body
	content, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		log.Println("[-] cannot read body:", err)
		c.JSON(http.StatusInternalServerError, "body error")
		return
	}

	// decode body
	var payload CommandMessage
	err = json.Unmarshal(content, &payload)

	if err != nil {
		log.Println("[-] cannot read json:", err)
		c.JSON(http.StatusInternalServerError, "json error")
		return
	}

	log.Printf("[+] payload: jobid: %s\n", payload.Id)

	// push body to redis
	log.Printf("[+] message destination [%s]\n", payload.Id)

	// push message to client queue
	_, err = db.Do("RPUSH", fmt.Sprintf("cmds_queue_%s", payload.Id), content)

	c.JSON(http.StatusOK, "ok")
}

func stats(c *gin.Context) {
	gid := c.Param("gid")
	nid := c.Param("nid")

	log.Printf("[+] gin: stats (gid: %s, nid: %s)\n", gid, nid)

	// read body
	content, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		log.Println("[-] cannot read body:", err)
		c.JSON(http.StatusInternalServerError, "body error")
		return
	}

	// decode body
	var payload StatsRequest
	err = json.Unmarshal(content, &payload)

	if err != nil {
		log.Println("[-] cannot read json:", err)
		c.JSON(http.StatusInternalServerError, "json error")
		return
	}

	// building Influxdb requests
	con, err := client.NewClient(&client.ClientConfig{
		Username: settings.Influxdb.User,
		Password: settings.Influxdb.Password,
		Database: settings.Influxdb.Db,
		Host:     settings.Influxdb.Host,
	})

	if err != nil {
		log.Println(err)
	}

	var timestamp = payload.Timestamp
	seriesList := make([]*client.Series, len(payload.Series))

	for i := 0; i < len(payload.Series); i++ {
		series := &client.Series{
			Name:    payload.Series[i][0].(string),
			Columns: []string{"time", "value"},
			// FIXME: add all points then write once
			Points: [][]interface{}{{
				timestamp * 1000, //influx expects time in ms
				payload.Series[i][1],
			}},
		}

		seriesList[i] = series
	}

	if err := con.WriteSeries(seriesList); err != nil {
		log.Println(err)
		return
	}

	c.JSON(http.StatusOK, "ok")
}

func event(c *gin.Context) {
	gid := c.Param("gid")
	nid := c.Param("nid")

	log.Printf("[+] gin: event (gid: %s, nid: %s)\n", gid, nid)

	content, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		log.Println("[-] cannot read body:", err)
		c.JSON(http.StatusInternalServerError, "body error")
		return
	}

	var payload EvenRequest
	log.Printf("%s", content)
	err = json.Unmarshal(content, &payload)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, "Error")
	}

	cmd := exec.Command(settings.Handlers.Binary,
		fmt.Sprintf("%s.py", payload.Name), gid, nid)

	cmd.Dir = settings.Handlers.Cwd
	//build env string
	var env []string
	if len(settings.Handlers.Env) > 0 {
		env = make([]string, 0, len(settings.Handlers.Env))
		for ek, ev := range settings.Handlers.Env {
			env = append(env, fmt.Sprintf("%v=%v", ek, ev))
		}

	}

	cmd.Env = env

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Println("Failed to open process stderr", err)
	}

	log.Println("Starting handler for", payload.Name, "event, for agent", gid, nid)
	err = cmd.Start()
	if err != nil {
		log.Println(err)
	} else {
		go func() {
			//wait for command to exit.
			cmderrors, err := ioutil.ReadAll(stderr)
			if len(cmderrors) > 0 {
				log.Printf("%s(%s:%s): %s", payload.Name, gid, nid, cmderrors)
			}

			err = cmd.Wait()
			if err != nil {
				log.Println("Failed to handle ", payload.Name, " event for agent: ", gid, nid, err)
			}
		}()
	}

	c.JSON(http.StatusOK, "ok")
}

func hubbleProxy(context *gin.Context) {
	hubble.ProxyHandler(context.Writer, context.Request)
}

var settings Settings

func main() {
	var cfg string
	var help bool

	flag.BoolVar(&help, "h", false, "Print this help screen")
	flag.StringVar(&cfg, "c", "", "Path to config file")
	flag.Parse()

	printHelp := func() {
		log.Println("agentcontroller [options]")
		flag.PrintDefaults()
	}

	if help {
		printHelp()
		return
	}

	if cfg == "" {
		log.Println("Missing required option -c")
		flag.PrintDefaults()
		os.Exit(1)
	}

	LoadTomlFile(cfg, &settings)

	log.Printf("[+] webservice: <%s>\n", settings.Main.Listen)
	log.Printf("[+] redis server: <%s>\n", settings.Main.RedisHost)

	pool = newPool(settings.Main.RedisHost, settings.Main.RedisPassword)
	router := gin.Default()

	go cmdreader()

	hubbleAuth.Install(hubbleAuth.NewAcceptAllModule())
	router.GET("/:gid/:nid/cmd", cmd)
	router.POST("/:gid/:nid/log", logs)
	router.POST("/:gid/:nid/result", result)
	router.POST("/:gid/:nid/stats", stats)
	router.POST("/:gid/:nid/event", event)
	router.GET("/:gid/:nid/hubble", hubbleProxy)
	// router.Static("/doc", "./doc")

	router.Run(settings.Main.Listen)
}
