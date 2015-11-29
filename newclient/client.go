package newclient
import (
	"github.com/Jumpscale/agentcontroller2/core"
	"github.com/Jumpscale/agentcontroller2/newclient/commandfactory"
	"fmt"
)

// A high-level client with future-based APIs for speaking to AgentController2
type Client struct{LowLevelClient}

type CommandTarget commandfactory.CommandTarget

func NewClient(address, redisPassword string) Client {
	return Client{NewLowLevelClient(address, redisPassword)}
}

func AnyNode() CommandTarget {
	return CommandTarget{}
}

func AllNodes() CommandTarget {
	return CommandTarget{Fanout: true}
}

// Retrieves information about the current live agents
func (client Client) LiveAgents() (<- chan []core.AgentID, <- chan error) {

	errChan := make(chan error)
	agentsChan := make(chan []core.AgentID)

	responses := DoneResponses(client.LowLevelClient.Execute(commandfactory.CommandInternalListAgents()))

	go func() {
		select {
		case response := <-responses:
			if response.Content.State == core.CommandStateError {
				errChan <- fmt.Errorf(response.Content.Data)
			} else {
				agentsChan <- parseCommandInternalListAgents(&response)
			}
		}

		close(errChan)
		close(agentsChan)
	}()

	return agentsChan, errChan
}


func (client Client) ExecuteExecutable(target commandfactory.CommandTarget,
	executable string, args []string) (<-chan ExecutableResult, <-chan error) {

	errChan := make(chan error)
	responseChan := make(chan ExecutableResult)

	command := commandfactory.CommandExecute(target, executable, args)
	responses := DoneResponses(client.LowLevelClient.Execute(command))

	go func() {
		select {
		case response := <-responses:
			if response.Content.State == core.CommandStateError {
				errChan <- fmt.Errorf(response.Content.Data)
			} else {
				responseChan <- parseCommandExecute(&response)
			}
		}

		close(errChan)
		close(responseChan)
	}()

	return responseChan, errChan
}