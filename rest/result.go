package rest

import (
	"encoding/json"
	"github.com/Jumpscale/agentcontroller2/core"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
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
	var payload core.CommandResult
	err = json.Unmarshal(content, &payload)

	if err != nil {
		log.Println("[-] cannot read json:", err)
		c.JSON(http.StatusInternalServerError, "json error")
		return
	}

	log.Println("Jobresult:", payload.ID)

	err = r.commandResponder(&payload)
	if err != nil {
		log.Println("Failed queue results")
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, "ok")
}
