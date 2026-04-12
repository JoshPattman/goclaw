package messages

import (
	"goclaw/agent"

	"github.com/JoshPattman/jpf"
)

type Message interface {
	Role() jpf.Role
	Content() agent.JsonObject
}

type ShrinkableMessage interface {
	Message
	Shrunk() Message
	IsShrunk() bool
}
