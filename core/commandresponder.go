package core

type CommandResponder interface {

	// Signals to the outside word that this message as been picked up
	SignalAsPickedUp(*Command)

	RespondToCommand(*CommandResponse) error
}