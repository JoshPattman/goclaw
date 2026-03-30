package messages

import (
	"goclaw/agent"

	"github.com/JoshPattman/jpf"
)

func NeedToExplicitlyStopMessage() Message {
	return needToEndMessage{}
}

type needToEndMessage struct{}

func (needToEndMessage) Role() jpf.Role {
	return jpf.UserRole
}

func (needToEndMessage) Content() agent.JsonObject {
	return agent.JsonObject{
		"explanation": "You called no tools, however you will continue iterating (calling no tools is not a useful thing to do). To stop iterating, please call the end_iteration tool by itself with no args.",
	}
}
