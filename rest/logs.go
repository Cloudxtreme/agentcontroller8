package rest
import (
	"github.com/gin-gonic/gin"
	"log"
	"io/ioutil"
	"net/http"
	"github.com/Jumpscale/agentcontroller8/utils"
)


func (r *Manager) logs(c *gin.Context) {
	agentID := utils.GetAgentID(c)

	log.Printf("[+] gin: log (%v)\n", agentID)

	// read body
	entry, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		log.Println("[-] cannot read body:", err)
		c.JSON(http.StatusInternalServerError, "error")
		return
	}

	err = r.agentLog.Push(agentID, entry)
	if err != nil {
		log.Println("Failed to store agent log entry")
		c.JSON(http.StatusInternalServerError, "error")
		return
	}

	c.JSON(http.StatusOK, "ok")
}