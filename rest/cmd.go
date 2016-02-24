package rest
import (
"github.com/gin-gonic/gin"
"log"
"net/http"
	"time"
"github.com/Jumpscale/agentcontroller8/core"
	"github.com/Jumpscale/agentcontroller8/utils"
)

func extractRoles(ctx *gin.Context) []core.AgentRole {
	var roles []core.AgentRole
	for _, role := range ctx.Request.URL.Query()["role"] {
		roles = append(roles, core.AgentRole(role))
	}
	return roles
}

func (r *Manager) cmd(c *gin.Context) {
	agentID := utils.GetAgentID(c)

	roles := extractRoles(c)
	log.Printf("[+] gin: execute (%v)\n", agentID)

	// listen for http closing
	notify := c.Writer.(http.CloseNotifier).CloseNotify()

	timeout := 60 * time.Second

	producer := r.producerChanFactory(agentID)

	data := &core.PollData{
		Roles:   roles,
		MsgChan: make(chan string),
	}

	select {
	case producer <- data:
	case <-time.After(timeout):
		c.String(http.StatusOK, "")
		return
	}
	//at this point we are sure this is the ONLY agent polling on /gid/nid/cmd

	var payload string

	select {
	case payload = <-data.MsgChan:
	case <-notify:
	case <-time.After(timeout):
	}

	c.String(http.StatusOK, payload)
}
