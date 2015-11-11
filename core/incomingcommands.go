package core


// A queue of incoming commands
type IncomingCommands interface {

	Pop() (*Command, error)

	Push(*Command) error
}