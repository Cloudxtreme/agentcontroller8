package redisdata_test
import (
	"github.com/garyburd/redigo/redis"
	"testing"
	"syscall"
	"github.com/stretchr/testify/assert"
)

func TestingRedisPool(t *testing.T) *redis.Pool {

	redisPort, redisPortSet := syscall.Getenv("TEST_REDIS_PORT")
	if !redisPortSet {
		redisPort = "6379"
	}

	pool := &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialTimeout("tcp", "localhost:" + redisPort, 0, 0, 0)

			if err != nil {
				panic(err.Error())
			}

			return c, err
		},
	}

	db := pool.Get()
	defer db.Close()

	_, err := db.Do("PING")
	assert.NoError(t, err)

	return pool
}

