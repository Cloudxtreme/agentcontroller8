package core

// Temporary storage for results of executed commands
type LoggedCommandResults interface {

	Push(commandResult *CommandResponse) error

	Pop() (*CommandResponse, error)
}