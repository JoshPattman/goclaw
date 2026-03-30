package messages

import (
	"goclaw/agent"

	"github.com/JoshPattman/jpf"
)

func ToolResponseMessage(responses []string) Message {
	return toolResponseMessage{responses}
}

type toolResponseMessage struct {
	responses []string
}

func (m toolResponseMessage) Role() jpf.Role {
	return jpf.UserRole
}

func (m toolResponseMessage) Content() agent.JsonObject {
	return agent.JsonObject{
		"explanation":    "Here are the responses from your tool calls",
		"tool_responses": m.responses,
	}
}
