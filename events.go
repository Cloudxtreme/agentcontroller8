package main

import (
	"encoding/json"
	"fmt"
	"github.com/Jumpscale/pygo"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

type EventsHandler struct {
	module pygo.Pygo
}

//EventRequest event request
type EventRequest struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func NewEventsHandler(settings *Events) (*EventsHandler, error) {
	opts := pygo.PyOpts{
		PythonPath: settings.PythonPath,
		Env: []string{
			fmt.Sprintf("HOME=%s", os.Getenv("HOME")),
		},
	}

	module, err := pygo.NewPy(settings.Module, &opts)
	if err != nil {
		return nil, err
	}

	handler := &EventsHandler{
		module: module,
	}
	log.Println("Calling handlers init")
	_, err = handler.module.Call("init", settings.Settings)
	if err != nil {
		return nil, err
	}
	log.Println("Init passed successfully")

	return handler, nil
}

func (handler *EventsHandler) Event(c *gin.Context) {
	gid := c.Param("gid")
	nid := c.Param("nid")

	log.Printf("[+] gin: event (gid: %s, nid: %s)\n", gid, nid)

	//force initializing of producer since the event is the first thing agent sends

	getProducerChan(gid, nid)

	content, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		log.Println("[-] cannot read body:", err)
		c.JSON(http.StatusInternalServerError, "body error")
		return
	}

	var payload EventRequest
	log.Printf("%s", content)
	err = json.Unmarshal(content, &payload)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, "Error")
	}

	igid, _ := strconv.Atoi(gid)
	inid, _ := strconv.Atoi(nid)

	go func(payload EventRequest, gid int, nid int) {
		_, err = handler.module.Apply(payload.Name, map[string]interface{}{
			"gid": gid,
			"nid": nid,
		})

		if err != nil {
			log.Println("Failed to handle ", payload.Name, " event for agent: ", gid, nid, err)
			log.Println(err, handler.module.Error())
		}

	}(payload, igid, inid)

	c.JSON(http.StatusOK, "ok")
}
