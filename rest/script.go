package rest

import (
	"github.com/Jumpscale/agentcontroller8/core"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

//Gets hashed scripts from redis.
func (r *Manager) script(c *gin.Context) {
	query := c.Request.URL.Query()
	hashes, ok := query["hash"]
	if !ok {
		// that's an error. Hash is required.
		c.String(http.StatusBadRequest, "Missing 'hash' param")
		return
	}

	id := hashes[0]

	jumpscriptContent, err := r.jumpscriptStore.Get(core.JumpScriptID(id))
	if err != nil {
		log.Println("Script get error:", err)
		c.String(http.StatusNotFound, "Script with hash '%s' not found", id)
		return
	}

	c.String(http.StatusOK, string(jumpscriptContent))
}
