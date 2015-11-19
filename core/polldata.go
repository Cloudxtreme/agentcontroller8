package core

// PollData Gets a chain for the caller to wait on, we return a chan chan string instead
// of chan string directly to make sure of the following:
// 1- The redis pop loop will not try to pop jobs out of the queue until there is a caller waiting
//    for new commands
// 2- Prevent multiple clients polling on a single gid:nid at the same time.
type PollData struct {
	Roles   []AgentRole
	MsgChan chan string
}

type ProducerChanFactory func(AgentID) chan<- *PollData
