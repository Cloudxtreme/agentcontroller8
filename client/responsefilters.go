package client

import "github.com/Jumpscale/agentcontroller8/core"

// Filters responses and only passes through the terminal ones
// May return more than one terminal response if more than one agent were responding
func TerminalResponses(incoming <-chan core.CommandResponse) <-chan core.CommandResponse {

	outgoing := make(chan core.CommandResponse)

	go func() {
		defer close(outgoing)
		for {
			select {
			case response, isOpen := <-incoming:
				if !isOpen {
					return
				}
				if core.IsTerminalCommandState(response.Content.State) {
					outgoing <- response
				}
			}
		}
	}()

	return outgoing
}
