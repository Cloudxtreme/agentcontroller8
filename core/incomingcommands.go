package core


// A queue of incoming commands
type CommandSource interface {

	Pop() (*Command, error)

	Push(*Command) error
}