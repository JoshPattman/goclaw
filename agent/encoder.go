package agent

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/JoshPattman/jpf"
)

type EncoderInput struct {
	Messages      []Message
	ToolDefs      []ToolDef
	Time          time.Time
	WorkingMemory string
}

func Encoder(systemPrompt JsonObject) jpf.Encoder[EncoderInput] {
	return &msgEncoder{
		systemPrompt,
	}
}

type msgEncoder struct {
	systemPrompt JsonObject
}

func (e *msgEncoder) BuildInputMessages(input EncoderInput) ([]jpf.Message, error) {
	messages := make([]jpf.Message, 0)

	messages = append(messages, jpf.Message{
		Role:    jpf.SystemRole,
		Content: objectContent(e.systemPrompt),
	})

	for _, msg := range input.Messages {
		messages = append(messages, jpf.Message{
			Role:    msg.Role(),
			Content: objectContent(msg.Content()),
		})
	}
	messages = append(messages, jpf.Message{
		Role:    jpf.UserRole,
		Content: objectContent(e.activeState(input)),
	})

	return messages, nil
}

func (e *msgEncoder) activeState(input EncoderInput) JsonObject {
	return JsonObject{
		"description":      "This is a current state message. You will have been provided with them at previous points in the conversation too, however they have been removed for brevity. This state message is currently up-to-date and active.",
		"active_tools":     input.ToolDefs,
		"current_datetime": input.Time.Format(time.RFC1123),
		"working_memory":   input.WorkingMemory,
	}
}

func objectContent(obj JsonObject) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "    ")
	err := enc.Encode(obj)
	if err != nil {
		panic(err)
	}
	return buf.String()
}
