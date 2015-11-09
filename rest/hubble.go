package rest
import "github.com/gin-gonic/gin"
import hublleProxy "github.com/Jumpscale/hubble/proxy"

func (r *Manager) handlHubbleProxy(context *gin.Context) {
	hublleProxy.ProxyHandler(context.Writer, context.Request)
}