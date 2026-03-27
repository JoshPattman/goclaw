package agent

import (
	"bytes"
	"encoding/json"

	"github.com/JoshPattman/jpf"
)

type message interface {
	Message() jpf.Message
}

type eventsMessage struct {
	events []Event
}

type eventsLLMObject struct {
	Preamble string  `json:"preamble"`
	Events   []Event `json:"events"`
}

func (msg eventsMessage) Message() jpf.Message {
	obj := eventsLLMObject{
		"In the time since your last message, the following events have occured. They may or may not require to to perform actions to handle them.",
		msg.events,
	}

	content, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		panic(err)
	}

	return jpf.Message{
		Role:    jpf.UserRole,
		Content: string(content),
	}
}

type toolCallArgs struct {
	ArgName string `json:"arg_name"`
	Value   any    `json:"value"`
}

type toolCall struct {
	ToolName string         `json:"tool_name"`
	Args     []toolCallArgs `json:"args"`
}

type toolCallsMessage struct {
	ToolCalls []toolCall `json:"tool_calls"`
}

func (msg toolCallsMessage) Message() jpf.Message {
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	enc.Encode(msg)

	return jpf.Message{
		Role:    jpf.AssistantRole,
		Content: buf.String(),
	}
}

type topolResponseMessage struct {
	Responses []string
}

type toolResponseObj struct {
	Preamble      string   `json:"preamble"`
	ToolResponses []string `json:"tool_responses"`
}

func (msg topolResponseMessage) Message() jpf.Message {
	obj := toolResponseObj{
		"The following are responses from the tools you called, in the same order you called them",
		nil,
	}
	for _, r := range msg.Responses {
		obj.ToolResponses = append(obj.ToolResponses, r)
	}
	content, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		panic(err)
	}
	return jpf.Message{
		Role:    jpf.AssistantRole,
		Content: string(content),
	}
}

type msgEncoder struct {
	systemPrompt string
}

func (e *msgEncoder) BuildInputMessages(input []message) ([]jpf.Message, error) {
	messages := make([]jpf.Message, 0, len(input)+1)

	messages = append(messages, jpf.Message{
		Role:    jpf.SystemRole,
		Content: e.systemPrompt,
	})

	for _, msg := range input {
		messages = append(messages, msg.Message())
	}

	return messages, nil
}
