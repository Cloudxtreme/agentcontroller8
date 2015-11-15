package rest

import (
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"github.com/Jumpscale/agentcontroller2/core"
)

func (r *Manager) result(c *gin.Context) {
	gid := c.Param("gid")
	nid := c.Param("nid")

	db := r.redisPool.Get()
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
