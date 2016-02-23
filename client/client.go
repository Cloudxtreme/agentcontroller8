package client

import (
	"fmt"
	"github.com/Jumpscale/agentcontroller8/client/commandfactory"
	"github.com/Jumpscale/agentcontroller8/client/responseparsing"
	"github.com/Jumpscale/agentcontroller8/core"
	"github.com/Jumpscale/agentcontroller8/scheduling"
)

// A high-level client with future-based APIs for speaking to AgentController8
type Client struct{ LowLevelClient }

func NewClient(address, redisPassword string) Client {
	return Client{NewLowLevelClient(address, redisPassword)}
}

func AnyNode() commandfactory.CommandTarget {
	return commandfactory.CommandTarget{Roles: []core.AgentRole{"*"}}
}

func AllNodes() commandfactory.CommandTarget {
	return commandfactory.CommandTarget{Fanout: true}
}

// Retrieves information about the current live agents
// Returns two buffered channels of size 1. Either of the two channels will be populated with a value for you.
func (client Client) LiveAgents() (<-chan []core.AgentID, <-chan error) {

	// Only 1 response expected

	errChan := make(chan error, 1)
	agentsChan := make(chan []core.AgentID, 1)

	responses := TerminalResponses(client.LowLevelClient.Execute(commandfactory.CommandInternalListAgents()))

	go func() {
		select {
		case response := <-responses:
			if response.Content.State == core.CommandStateError {
				errChan <- fmt.Errorf(response.Content.Data)
			} else {
				agentsChan <- responseparsing.InternalListAgents(&response)
			}
		}

		close(errChan)
		close(agentsChan)
	}()

	return agentsChan, errChan
}


// Returns two buffered channels of size 1. Either of the two channels will be populated with a value for you.
func (client Client) ExecuteExecutable(target commandfactory.CommandTarget,
	executable string, args []string) (<-chan []responseparsing.ExecutableResult, <-chan []error) {

	// Expecting as many responses as there are targeted agents

  errChan := make(chan []error, 1)
  responseChan := make(chan []responseparsing.ExecutableResult, 1)	// Must be 1-buffered

  go func() {
    defer close(errChan)
    defer close(responseChan)

    command := commandfactory.CommandExecute(target, executable, args)
    responses := exhaust(TerminalResponses(client.LowLevelClient.Execute(command)))

    results := []responseparsing.ExecutableResult{}
    errors := []error{}

    for _, response := range responses {
      if response.Content.State == core.CommandStateError {
        errors = append(errors, fmt.Errorf(response.Content.Data))
      } else {
        results = append(results, responseparsing.Execute(&response))
      }
    }

    responseChan <- results
    if len(errors) > 0 {
      errChan <- errors
    }
  }()

  return responseChan, errChan
}


// Returns two buffered channels of size 1. Either of the two channels will be populated with a value for you.
func (client Client) GetProcessStats(target commandfactory.CommandTarget) (<-chan [][]responseparsing.RunningCommandStats, <-chan []error) {

	// Expecting as many responses as there are targeted agents

	errChan := make(chan []error, 1)
	responseChan := make(chan [][]responseparsing.RunningCommandStats, 1)

	go func() {
		defer close(errChan)
		defer close(responseChan)

    command := commandfactory.CommandGetProcessStats(target)
    responses := exhaust(TerminalResponses(client.LowLevelClient.Execute(command)))

    results := [][]responseparsing.RunningCommandStats{}
    errors := []error{}

    for _, response := range responses {
     if response.Content.State == core.CommandStateError {
       errors = append(errors, fmt.Errorf(response.Content.Data))
      } else {
       results = append(results, responseparsing.GetProcessStats(&response))
      }
    }

    responseChan <- results
    if len(errors) > 0 {
      errChan <- errors
    }
	}()

	return responseChan, errChan
}


// Returns two buffered channels of size 1. Either of the two channels will be populated with a value for you.
func (client Client) SchedulerListJobs() (<-chan []scheduling.Job, <-chan error) {

	// Only 1 response expected

	errChan := make(chan error, 1)
	responseChan := make(chan []scheduling.Job, 1)

	go func() {
		defer close(errChan)
		defer close(responseChan)

   	command := commandfactory.CommandInternalSchedulerListJobs()
    responses := exhaust(TerminalResponses(client.LowLevelClient.Execute(command)))

    if len(responses) == 0 {
      errChan <- fmt.Errorf("Did not receive a response")
      return
    }

    response := responses[0]
    if response.Content.State == core.CommandStateError {
      errChan <- fmt.Errorf(response.Content.Data)
    } else {
      responseChan <- responseparsing.InternalSchedulerListJobs(&response)
    }
	}()

	return responseChan, errChan
}

// The channel of scheduling.Job may return nothing and immediately be closed if there are no jobs with
// the specified ID.
// Returns two buffered channels of size 1. Either of the two channels will be populated with a value for you.
func (client Client) SchedulerGetJob(id string) (<-chan scheduling.Job, <-chan error) {

	jobChan := make(chan scheduling.Job, 1)
	jobsChan, errChan := client.SchedulerListJobs()

	go func() {
    jobs := <-jobsChan
    for _, job := range jobs {
      if job.ID == id {
        jobChan <- job
      }
    }
    close(jobChan)
  }()

	return jobChan, errChan
}


func (client Client) SchedulerAddJob(id string, scheduledCommand *core.Command, timingSpec string) <-chan error {

	// Only 1 response expected

	errChan := make(chan error, 1)

	go func() {
		defer close(errChan)

   	command := commandfactory.CommandInternalSchedulerAdd(id, scheduledCommand, timingSpec)
    responses := exhaust(TerminalResponses(client.LowLevelClient.Execute(command)))

    if len(responses) == 0 {
      return
    }

    response := responses[0]

    if response.Content.State == core.CommandStateError {
      errChan <- fmt.Errorf(responses[0].Content.Data)
    }
	}()

	return errChan
}


// Returns two buffered channels of size 1. Either of the two channels will be populated with a value for you.
func (client Client) SchedulerRemoveJob(id string) (chan bool, <-chan error) {

	// Only 1 response expected

	errChan := make(chan error, 1)
	responseChan := make(chan bool, 1)

	go func() {
		defer close(errChan)
		defer close(responseChan)

    command := commandfactory.CommandInternalSchedulerRemoveJob(id)
    response := exhaust(TerminalResponses(client.LowLevelClient.Execute(command)))[0]

    if response.Content.State == core.CommandStateError {
      errChan <- fmt.Errorf(response.Content.Data)
    } else {
      responseChan <- responseparsing.InternalSchedulerRemoveJob(&response)
    }
	}()

	return responseChan, errChan
}
