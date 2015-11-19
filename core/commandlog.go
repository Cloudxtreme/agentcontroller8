package core


// Temporary storage for executed commands
type CommandLog interface {

	Push(*Command) error

	BlockingPop() (*Command, error)
}