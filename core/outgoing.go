package core

type Outgoing interface {

	// Signals to the outside word that this message as been queued
	SignalAsQueued(*Command)

	RespondToCommand(*CommandResult) error
}