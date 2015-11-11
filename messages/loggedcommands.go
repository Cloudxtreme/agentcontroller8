package messages

// Temporary storage for executed commands
type LoggedCommands interface {

	Push(*CommandMessage) error

	Pop() (*CommandMessage, error)
}