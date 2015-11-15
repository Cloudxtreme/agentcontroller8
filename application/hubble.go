package application
import (
	"time"
	"log"
	"fmt"
	hubbleAgent "github.com/Jumpscale/hubble/agent"
)

//StartSyncthingHubbleAgent start the builtin hubble agent required for Syncthing
func startHubbleAgent(hubblePort int) {

	wsURL := fmt.Sprintf("ws://127.0.0.1:%d/0/0/hubble", hubblePort)
	log.Println("Starting local hubble agent at", wsURL)
	agent := hubbleAgent.NewAgent(wsURL, "controller", "", nil)
	var onExit func(agt hubbleAgent.Agent, err error)

	onExit = func(agt hubbleAgent.Agent, err error) {
		if err != nil {
			go func() {
				time.Sleep(3 * time.Second)
				agt.Start(onExit)
			}()
		}
	}

	agent.Start(onExit)
}