package messages

import (
	"goclaw/agent"

	"github.com/JoshPattman/jpf"
)

func ToolResponseMessage(responses []string) Message {
	return toolResponseMessage{responses, false}
}

type toolResponseMessage struct {
	responses []string
	shrink    bool
}

func (m toolResponseMessage) Role() jpf.Role {
	return jpf.UserRole
}

func (m toolResponseMessage) Content() agent.JsonObject {
	if m.shrink {
		return agent.JsonObject{
			"explanation": "This tool response has been hidden to save tokens. You should have written any useful info into your scratchpad.",
		}
	} else {
		return agent.JsonObject{
			"explanation":    "Here are the responses from your tool calls",
			"tool_responses": m.responses,
		}
	}
}

func (m toolResponseMessage) Shrunk() Message {
	m.shrink = true
	return m
}

func (m toolResponseMessage) IsShrunk() bool {
	return m.shrink
}
