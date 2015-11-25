package newclient
import (
"github.com/garyburd/redigo/redis"
"github.com/Jumpscale/agentcontroller2/core"
	"github.com/Jumpscale/agentcontroller2/redisdata/ds"
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
func (c Client) Execute(command *core.Command) (<- chan core.CommandResponse, error) {
	// TODO
	panic("TODO")
}