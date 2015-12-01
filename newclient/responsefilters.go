package newclient
import "github.com/Jumpscale/agentcontroller2/core"


// Filters responses and only passes through the terminal response
func TerminalResponse(incoming <-chan core.CommandResponse) <-chan core.CommandResponse {

	outgoing := make(chan core.CommandResponse)

	go func() {
		defer close(outgoing)
		for {
			select {
			case response, isOpen := <-incoming:
				if !isOpen {
					close(outgoing)
					return
				}
				if core.IsTerminalCommandState(response.Content.State) {
					outgoing <- response
					return
				}
			}
		}
	}()

	return outgoing
}