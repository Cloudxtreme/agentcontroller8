package rest

import (
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"github.com/Jumpscale/agentcontroller8/core"
	"github.com/Jumpscale/agentcontroller8/utils"
)

func (r *Manager) result(c *gin.Context) {
	agentID := utils.GetAgentID(c)

	log.Printf("[+] gin: result (%v)\n", agentID)

	// read body
	content, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		log.Println("[-] cannot read body:", err)
		c.JSON(http.StatusInternalServerError, "body error")
		return
	}

	// decode body
	commandResult, err := core.CommandResponseFromJSON(content)

	if err != nil {
		log.Println("[-] cannot read json:", err)
		c.JSON(http.StatusInternalServerError, "json error")
		return
	}

	log.Println("Jobresult:", commandResult.Content.ID)

	r.commandResponder.RespondToCommand(commandResult)

	c.JSON(http.StatusOK, "ok")
}
