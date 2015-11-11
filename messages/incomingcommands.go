package messages

// A queue of incoming commands
type IncomingCommands interface {

	Pop() (*CommandMessage, error)

	Push(*CommandMessage) error
}