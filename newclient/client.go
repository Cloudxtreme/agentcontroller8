package newclient
import (
	"github.com/Jumpscale/agentcontroller2/core"
	"github.com/Jumpscale/agentcontroller2/newclient/commandfactory"
	"fmt"
	"encoding/json"
	"strings"
	"github.com/Jumpscale/agentcontroller2/utils"
)

type Client struct {LowLevelClient}

// Retrieves information about the current live agents
func (client Client) LiveAgents() (<- chan []core.AgentID, <- chan error) {

	errChan := make(chan error)
	agentsChan := make(chan []core.AgentID)

	responses := DoneResponses(client.LowLevelClient.Execute(commandfactory.CommandInternalListAgents()))

	go func() {
		select {
		case response := <- responses:
			if response.Content.State == core.CommandStateError {
				errChan <- fmt.Errorf(response.Content.Data)
			} else {
				data := response.Content.Data
				agentMap := map[string](interface{}){}
				err := json.Unmarshal([]byte(data), &agentMap)
				if err != nil {
					panic(fmt.Errorf("Malformed response"))
				}

				var agents []core.AgentID
				for agentStr, _ := range agentMap {
					gidnid := strings.Split(agentStr, ":")
					if len(gidnid) != 2 {
						panic("Malformed response")
					}
					agents = append(agents, utils.AgentIDFromStrings(gidnid[0], gidnid[1]))
				}

				agentsChan <- agents
			}
		}

		close(errChan)
		close(agentsChan)
	}()

	return agentsChan, errChan
}