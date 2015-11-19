package utils
import (
	"fmt"
	"github.com/Jumpscale/agentcontroller2/core"
	"github.com/gin-gonic/gin"
	"crypto/md5"
	"encoding/json"
)

// Extracts an Agent ID from a Gin context
func GetAgentID(ctx *gin.Context) core.AgentID {
	var agentID core.AgentID
	fmt.Sscanf(ctx.Param("gid"), "%v", &agentID.GID)
	fmt.Sscanf(ctx.Param("nid"), "%v", &agentID.NID)
	return agentID
}

func MD5Hex(data []byte) string {
	return fmt.Sprintf("%x", md5.Sum(data))
}

func MustJsonMarshal(whateves interface{}) []byte {
	data, err := json.Marshal(whateves)
	if err != nil {
		panic(err)
	}
	return data
}

func MustJsonUnmarshal(data []byte, target interface{}) {
	err := json.Unmarshal(data, target)
	if err != nil {
		panic(err)
	}
}