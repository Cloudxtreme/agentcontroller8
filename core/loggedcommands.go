package core


// Temporary storage for executed commands
type CommandLog interface {

	Push(*Command) error

	Pop() (*Command, error)
}