package client
import "github.com/Jumpscale/agentcontroller8/core"

// Blocks and exhausts the given channel and returns all the received values in a single array
func exhaust(respChan <-chan core.CommandResponse) []core.CommandResponse {
  responses := []core.CommandResponse{}
  for {
    select {
    case response, isOpen := <- respChan:
      if !isOpen {return responses}
      responses = append(responses, response)
    }
  }
}
