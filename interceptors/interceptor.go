package interceptors

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"github.com/garyburd/redigo/redis"
)

const (
	scriptHashTimeout = 86400 // seconds
)

type commandInterceptor func(map[string]interface{}, *Manager) (map[string]interface{}, error)

type Manager struct {
	redisPool *redis.Pool
	interceptors map[string]commandInterceptor
}

func NewManager(redisPool *redis.Pool) *Manager {
	return &Manager {
		redisPool: redisPool,
		interceptors: map[string]commandInterceptor{
			"jumpscript_content": jumpscriptHasherInterceptor,
		},
	}
}

// Hashes jumpscripts executed by the jumpscript_content and store it in redis. Alters the passed command as needed
func jumpscriptHasherInterceptor(cmd map[string]interface{}, manager *Manager) (map[string]interface{}, error) {
	datastr, ok := cmd["data"].(string)
	if !ok {
		return nil, errors.New("Expecting command 'data' to be string")
	}

	data := make(map[string]interface{})
	err := json.Unmarshal([]byte(datastr), &data)
	if err != nil {
		return nil, err
	}

	content, ok := data["content"]
	if !ok {
		return nil, errors.New("jumpscript_content doesn't have content payload")
	}

	contentstr, ok := content.(string)
	if !ok {
		return nil, errors.New("Expected 'content' to be string")
	}

	hash := fmt.Sprintf("%x", md5.Sum([]byte(contentstr)))

	db := manager.redisPool.Get()
	defer db.Close()

	if _, err := db.Do("SET", hash, contentstr, "EX", scriptHashTimeout); err != nil {
		return nil, err
	}

	//hash is stored. Now modify the command and forward it.
	delete(data, "content")
	data["hash"] = hash

	updatedDatastr, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	cmd["data"] = string(updatedDatastr)

	return cmd, nil
}

// InterceptCommand intercepts raw command data and manipulate it if needed.
func (manager *Manager) Intercept(command string) string {
	cmd := make(map[string]interface{})

	err := json.Unmarshal([]byte(command), &cmd)
	if err != nil {
		log.Println("Error: failed to intercept command", command, err)
	}

	cmdName, ok := cmd["cmd"].(string)
	if !ok {
		log.Println("Expected 'cmd' to be string")
		return command
	}

	interceptor, ok := manager.interceptors[cmdName]
	if ok {
		cmd, err = interceptor(cmd, manager)
		if err == nil {
			if data, err := json.Marshal(cmd); err == nil {
				command = string(data)
			} else {
				log.Println("Failed to serialize intercepted command", err)
			}
		} else {
			log.Println("Failed to intercept command", err)
		}
	}
	log.Println("Command:", command)
	return command
}
