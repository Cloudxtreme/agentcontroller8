package newclient
import (
"github.com/garyburd/redigo/redis"
"github.com/Jumpscale/agentcontroller2/core"
	"github.com/Jumpscale/agentcontroller2/redisdata/ds"
	"fmt"
	"time"
)

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

// Sends a command and returns a channel for reading the response
func (c Client) Execute(command *core.Command, timeout time.Duration) (<- chan core.CommandResponse, error) {

	err := c.Send(command)
	if err != nil {
		return nil, err
	}

	responseChan := make(chan core.CommandResponse)

	go func() {

		// Waiting for execution to finish
		waitedOnList := ds.GetList(fmt.Sprintf("cmd.%s.queued", command.Content.ID))
		data, err := waitedOnList.BlockingRightPopLeftPush(c.connPool, timeout, waitedOnList)
		if err != nil {
			panic(fmt.Sprintf("Redis error: %v", err))
		}

		if data == nil {
			// Timed-out
			close(responseChan)
			return
		}

		responses, err := ds.GetHash(fmt.Sprintf("jobresult:%s", command.Content.ID)).ToStringMap(c.connPool)
		if err != nil {
			panic(fmt.Sprintf("Redis error: %v", err))
		}

		for _, responseJSON := range responses {
			response, err := core.CommandResponseFromJSON([]byte(responseJSON))
			if err != nil {
				panic(fmt.Sprintf("Malformed response! %v", err))
			}
			responseChan <- *response
		}

		close(responseChan)
	}()

	return responseChan, nil
}