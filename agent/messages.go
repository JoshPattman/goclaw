package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/JoshPattman/jpf"
)

type message interface {
	Message() jpf.Message
}

type eventsMessage struct {
	events []Event
}

func eventToString(e Event) string {
	return fmt.Sprintf("Event of kind '%s':\n```event\n%s\n```", e.Kind, e.Content)
}

func (msg eventsMessage) Message() jpf.Message {
	eventStrings := make([]string, len(msg.events))
	for i, e := range msg.events {
		eventStrings[i] = eventToString(e)
	}

	content := fmt.Sprintf(
		"The following events have occured:\n\n%s",
		strings.Join(eventStrings, "\n\n"),
	)

	return jpf.Message{
		Role:    jpf.UserRole,
		Content: content,
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

func (msg topolResponseMessage) Message() jpf.Message {
	formatted := make([]string, len(msg.Responses))
	for i, r := range msg.Responses {
		formatted[i] = fmt.Sprintf("```tool_response\n%s\n```", r)
	}

	content := fmt.Sprintf(
		"The following tool responses were produced:\n\n%s",
		strings.Join(formatted, "\n\n"),
	)

	return jpf.Message{
		Role:    jpf.AssistantRole,
		Content: content,
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
