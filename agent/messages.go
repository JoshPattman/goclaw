package agent

import (
	"bytes"
	"encoding/json"

	"github.com/JoshPattman/jpf"
)

type JsonObject map[string]any

type Message interface {
	Role() jpf.Role
	Content() JsonObject
}

func EventsMessage(events ...Event) Message {
	return eventsMessage{events}
}

func ToolResponseMessage(responses []string) Message {
	return ToolResponseMessage(responses)
}

func EmptyToolCallsMessage() Message {
	return toolCallsMessage{}
}

func NeedToExplicitlyStopMessage() Message {
	return needToEndMessage{}
}

type eventsMessage struct {
	events []Event
}

func (eventsMessage) Role() jpf.Role {
	return jpf.UserRole
}

func (m eventsMessage) Content() JsonObject {
	eventObjects := make([]JsonObject, len(m.events))
	for i, e := range m.events {
		eventObjects[i] = JsonObject{
			"kind":    e.Kind(),
			"content": e.Content(),
		}
	}
	return JsonObject{
		"explanation": "Some events have occured since your last message. They might be relevant or you might be able to ignore them.",
		"events":      eventObjects,
	}
}

type toolCallsMessage struct {
	ToolCalls []toolCall `json:"tool_calls"`
}

type toolCallArgs struct {
	ArgName string `json:"arg_name"`
	Value   any    `json:"value"`
}

type toolCall struct {
	ToolName string         `json:"tool_name"`
	Args     []toolCallArgs `json:"args"`
}

func (m toolCallsMessage) Role() jpf.Role {
	return jpf.AssistantRole
}

func (m toolCallsMessage) Content() JsonObject {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(m)
	if err != nil {
		panic(err)
	}
	var result JsonObject
	err = json.NewDecoder(&buf).Decode(&result)
	if err != nil {
		panic(err)
	}
	return result
}

type toolResponseMessage struct {
	responses []string
}

func (m toolResponseMessage) Role() jpf.Role {
	return jpf.UserRole
}

func (m toolResponseMessage) Content() JsonObject {
	return JsonObject{
		"explanation":    "Here are the responses from your tool calls",
		"tool_responses": m.responses,
	}
}

type needToEndMessage struct{}

func (needToEndMessage) Role() jpf.Role {
	return jpf.UserRole
}

func (needToEndMessage) Content() JsonObject {
	return JsonObject{
		"explanation": "You called no tools, however you will continue iterating (calling no tools is not a useful thing to do). To stop iterating, please call the end_iteration tool by itself with no args.",
	}
}
