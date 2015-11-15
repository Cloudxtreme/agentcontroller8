package core

type CommandResponder interface {

	// Signals to the outside word that this message as been queued
	SignalAsQueued(*Command)

	RespondToCommand(*CommandResponse) error
}