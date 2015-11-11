package core


// Temporary storage for executed commands
type LoggedCommands interface {

	Push(*Command) error

	Pop() (*Command, error)
}