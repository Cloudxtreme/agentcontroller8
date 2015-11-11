package messages

type Outgoing interface {

	// Signals to the outside word that this message as been queued
	SignalAsQueued(*CommandMessage)

	RespondToCommand(*CommandResultMessage) error
}