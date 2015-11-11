package core

// Temporary storage for results of executed commands
type LoggedCommandResults interface {

	Push(commandResult *CommandResult) error

	Pop() (*CommandResult, error)
}