package messages

type OutgoingSignals interface {

	// Signals to the outside word that this message as been queued
	SignalAsQueued(command *CommandMessage)
}