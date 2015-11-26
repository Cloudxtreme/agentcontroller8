package newclient
import (
	"github.com/garyburd/redigo/redis"
	"github.com/Jumpscale/agentcontroller2/core"
	"github.com/Jumpscale/agentcontroller2/redisdata/ds"
	"fmt"
	"time"
	"strings"
	"github.com/Jumpscale/agentcontroller2/utils"
)

// The timeout used in each blocking response-receiving step. It is not the total timeout for receiving
// a response.
const responseTimeout = 1 * time.Minute

type Client struct {
	connPool *redis.Pool
}

func newRedisPool(address string, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", address)

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

func NewClient(redisAddress, redisPassword string) Client {
	return Client{
		connPool: newRedisPool(redisAddress, redisPassword),
	}
}

// Sends a command without receiving a response
func (c Client) Send(command *core.Command) error {
	return ds.GetCommandList("cmds.queue").RightPush(c.connPool, command)
}

// Blocks and then returns true when the command is picked up, returns false otherwise
func (c Client) isPickedUp(command *core.Command) bool {
	signal := ds.GetList(fmt.Sprintf("cmd.%s.queued", command.Content.ID))
	data, err := signal.BlockingRightPopLeftPush(c.connPool, responseTimeout, signal)
	if err != nil {
		panic(fmt.Errorf("Redis error: %v", err))
	}
	if data == nil {
		// Timed-out!
		return false
	}

	return true
}

// Reads the current responses for the specified command
func (c Client) responsesFor(command *core.Command) map[core.AgentID]core.CommandResponse {

	jsonResponses, err := ds.GetHash(fmt.Sprintf("jobresult:%s", command.Content.ID)).ToStringMap(c.connPool)
	if err != nil {
		panic(fmt.Sprintf("Redis error: %v", err))
	}

	responses := make(map[core.AgentID]core.CommandResponse)

	for strAgentID, jsonResponse := range jsonResponses {

		response, err := core.CommandResponseFromJSON([]byte(jsonResponse))
		if err != nil {
			panic(fmt.Sprintf("Malformed response! %v", err))
		}

		gidnid := strings.Split(strAgentID, ":")
		agentID := utils.AgentIDFromStrings(gidnid[0], gidnid[1])

		responses[agentID] = *response
	}

	return responses
}

// Blocks and returns true when the specified command on the specified Agent is done, false otherwise
func (c Client) isDone(command *core.Command, agentID core.AgentID) bool {
	name := fmt.Sprintf("cmd.%s.%d.%d", command.Content.ID, agentID.GID, agentID.NID)
	signal := ds.GetList(name)

	data, err := signal.BlockingRightPopLeftPush(c.connPool, responseTimeout, signal)
	if err != nil {
		panic(fmt.Errorf("Redis error: %v", err))
	}

	if data == nil {
		// Timed-out!
		return false
	}

	return true
}

// Sends a command and returns a channel for reading the responses
func (c Client) Execute(command *core.Command) <- chan core.CommandResponse {

	err := c.Send(command)
	if err != nil {
		panic(fmt.Errorf("Redis error: %v", err))
	}

	responseChan := make(chan core.CommandResponse)

	go func() {

		if !c.isPickedUp(command) {
			// Something is not right!
			close(responseChan)
			return
		}

		initialResponses := c.responsesFor(command)
		for _, response := range initialResponses {
			responseChan <- response
		}

		// Initial command state is the state of any of the initial responses
		var initialCommandState string
		{
			for _, response := range initialResponses {
				initialCommandState = response.Content.State
			}
		}

		if initialCommandState == core.CommandStateSuccess || initialCommandState == core.CommandStateError {
			// This is done!
			close(responseChan)
			return
		}

		for agentID, _ := range c.responsesFor(command) {
			if c.isDone(command, agentID) {
				responseChan <- c.responsesFor(command)[agentID]
			}
		}

		close(responseChan)
	}()

	return responseChan
}