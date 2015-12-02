package newclient
import (
	"github.com/Jumpscale/agentcontroller2/core"
	"github.com/Jumpscale/agentcontroller2/newclient/commandfactory"
	"fmt"
	"github.com/Jumpscale/agentcontroller2/scheduling"
)

// A high-level client with future-based APIs for speaking to AgentController2
type Client struct{LowLevelClient}

func NewClient(address, redisPassword string) Client {
	return Client{NewLowLevelClient(address, redisPassword)}
}

func AnyNode() commandfactory.CommandTarget {
	return commandfactory.CommandTarget{}
}

func AllNodes() commandfactory.CommandTarget {
	return commandfactory.CommandTarget{Fanout: true}
}

// Retrieves information about the current live agents
func (client Client) LiveAgents() (<- chan []core.AgentID, <- chan error) {

	errChan := make(chan error)
	agentsChan := make(chan []core.AgentID)

	responses := TerminalResponses(client.LowLevelClient.Execute(commandfactory.CommandInternalListAgents()))

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
	responses := TerminalResponses(client.LowLevelClient.Execute(command))

	go func() {
		defer close(errChan)
		defer close(responseChan)

		for {
			select {
			case response, isOpen := <-responses:
				if !isOpen { return }
				if response.Content.State == core.CommandStateError {
					errChan <- fmt.Errorf(response.Content.Data)
				} else {
					responseChan <- parseCommandExecute(&response)
				}
			}
		}
	}()

	return responseChan, errChan
}

func (client Client) GetProcessStats(target commandfactory.CommandTarget) (<-chan []RunningCommandStats, <-chan error) {

	errChan := make(chan error)
	responseChan := make(chan []RunningCommandStats)

	command := commandfactory.CommandGetProcessStats(target)
	responses := TerminalResponses(client.LowLevelClient.Execute(command))

	go func() {
		defer close(errChan)
		defer close(responseChan)

		for {
			select {
			case response, isOpen := <-responses:
				if !isOpen { return }
				if response.Content.State == core.CommandStateError {
					errChan <- fmt.Errorf(response.Content.Data)
				} else {
					responseChan <- parseCommandGetProcessStats(&response)
				}
			}
		}
	}()

	return responseChan, errChan
}

func (client Client) SchedulerListJobs() (<-chan []scheduling.Job, <-chan error) {

	errChan := make(chan error)
	responseChan := make(chan []scheduling.Job)

	command := commandfactory.CommandInternalSchedulerListJobs()
	responses := TerminalResponses(client.LowLevelClient.Execute(command))

	go func() {
		defer close(errChan)
		defer close(responseChan)

		for {
			select {
			case response, isOpen := <-responses:
				if !isOpen { return }
				if response.Content.State == core.CommandStateError {
					errChan <- fmt.Errorf(response.Content.Data)
				} else {
					responseChan <- parseCommandInternalSchedulerListJobs(&response)
				}
			}
		}
	}()

	return responseChan, errChan
}

func (client Client) SchedulerAddJob(id string, scheduledCommand *core.Command, timingSpec string) <-chan error {

	errChan := make(chan error)

	command := commandfactory.CommandInternalSchedulerAdd(id, scheduledCommand, timingSpec)
	responses := TerminalResponses(client.LowLevelClient.Execute(command))

	go func() {
		defer close(errChan)

		for {
			select {
			case response, isOpen := <-responses:
				if !isOpen { return }
				if response.Content.State == core.CommandStateError {
					errChan <- fmt.Errorf(response.Content.Data)
				}
			}
		}
	}()

	return errChan
}