package core

// Temporary storage for results of executed commands
type LoggedCommandResponses interface {

	Push(*CommandResponse) error

	Pop() (*CommandResponse, error)
}