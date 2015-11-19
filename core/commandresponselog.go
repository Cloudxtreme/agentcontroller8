package core

// Temporary storage for responses of executed commands
type CommandResponseLog interface {

	Push(*CommandResponse) error

	BlockingPop() (*CommandResponse, error)
}