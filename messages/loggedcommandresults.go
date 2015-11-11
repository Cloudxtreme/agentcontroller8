package messages

// Temporary storage for results of executed commands
type LoggedCommandResults interface {

	Push(commandResult *CommandResultMessage) error

	Pop() (*CommandResultMessage, error)
}