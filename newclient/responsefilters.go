package newclient
import "github.com/Jumpscale/agentcontroller2/core"


// Filters responses and only passes through the success/error final responses
func DoneResponses(incoming <-chan core.CommandResponse) <-chan core.CommandResponse {

	outgoing := make(chan core.CommandResponse)

	go func() {
		for {
			select {
			case response, isOpen := <-incoming:
				if !isOpen {
					close(outgoing)
					return
				}
				state := response.Content.State
				if state == core.CommandStateSuccess || state == core.CommandStateError {
					outgoing <- response
				}
			}
		}
	}()

	return outgoing
}