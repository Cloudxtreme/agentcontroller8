package messages

// Processed and logged commands
type LoggedCommands interface {

	Push(*CommandMessage) error

	Pop() (*CommandMessage, error)
}