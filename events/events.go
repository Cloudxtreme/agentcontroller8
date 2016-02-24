package events

import (
	"encoding/json"
	"fmt"
	"github.com/Jumpscale/agentcontroller8/configs"
	"github.com/Jumpscale/agentcontroller8/core"
	"github.com/Jumpscale/agentcontroller8/utils"
	"github.com/Jumpscale/pygo"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Handler struct {
	module           pygo.Pygo
	enabled          bool
	agentInformation core.AgentInformationStorage
}

//EventRequest event request
type EventRequest struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func NewEventsHandler(settings *configs.Extension, agentInformation core.AgentInformationStorage) (*Handler, error) {
	opts := pygo.PyOpts{
		PythonBinary: settings.GetPythonBinary(),
		PythonPath:   settings.PythonPath,
		Env: []string{
			fmt.Sprintf("HOME=%s", os.Getenv("HOME")),
		},
	}

	var module pygo.Pygo
	var err error
	if settings.Enabled {
		module, err = pygo.NewPy(settings.Module, &opts)
		if err != nil {
			return nil, err
		}

		log.Println("Calling handlers init")
		_, err = module.Call("init", settings.Settings)
		if err != nil {
			return nil, err
		}
		log.Println("Init passed successfully")
	}

	handler := &Handler{
		module:           module,
		enabled:          settings.Enabled,
		agentInformation: agentInformation,
	}

	return handler, nil
}

func (handler *Handler) Event(c *gin.Context) {
	if !handler.enabled {
		c.JSON(http.StatusOK, "ok")
		return
	}

	agentID := utils.GetAgentID(c)

	log.Printf("[+] gin: event (%v)\n", agentID)

	handler.agentInformation.MarkAsAlive(agentID)

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

	go func(payload EventRequest, gid int, nid int) {
		_, err = handler.module.Apply(payload.Name, map[string]interface{}{
			"gid": gid,
			"nid": nid,
		})

		if err != nil {
			log.Println("Failed to handle ", payload.Name, " event for agent: ", gid, nid, err)
			log.Println(err, handler.module.Error())
		}

	}(payload, int(agentID.GID), int(agentID.NID))

	c.JSON(http.StatusOK, "ok")
}
