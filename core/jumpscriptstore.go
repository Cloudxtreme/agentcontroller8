package core

type JumpScriptID string
type JumpScriptContent string

// A persistent storage of JumpScript scripts
type JumpScriptStore interface {

	Add(JumpScriptContent) (JumpScriptID, error)

	Get(JumpScriptID) (JumpScriptContent, error)
}