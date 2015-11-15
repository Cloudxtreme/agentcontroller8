package utils
import (
"fmt"
"github.com/Jumpscale/agentcontroller2/core"
	"github.com/gin-gonic/gin"
)

// Extracts an Agent ID from a Gin context
func GetAgentID(ctx *gin.Context) core.AgentID {
	var agentID core.AgentID
	fmt.Sscanf(ctx.Param("gid"), "%v", &agentID.GID)
	fmt.Sscanf(ctx.Param("nid"), "%v", &agentID.NID)
	return agentID
}
