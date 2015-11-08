package rest
import (
"github.com/gin-gonic/gin"
"log"
	"io/ioutil"
"net/http"
	"fmt"
)


func (r *RestInterface) logs(c *gin.Context) {
	gid := c.Param("gid")
	nid := c.Param("nid")

	db := r.redisPool.Get()
	defer db.Close()

	log.Printf("[+] gin: log (gid: %s, nid: %s)\n", gid, nid)

	// read body
	content, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		log.Println("[-] cannot read body:", err)
		c.JSON(http.StatusInternalServerError, "error")
		return
	}

	// push body to redis
	id := fmt.Sprintf("%s:%s:log", gid, nid)
	log.Printf("[+] message destination [%s]\n", id)

	// push message to client queue
	_, err = db.Do("RPUSH", id, content)

	c.JSON(http.StatusOK, "ok")
}