package core


// A queue of incoming commands
type CommandSource interface {

	BlockingPop() (*Command, error)

	Push(*Command) error
}