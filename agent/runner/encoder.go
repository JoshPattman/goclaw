package runner

import (
	"bytes"
	"encoding/json"
	"goclaw/agent"
	"goclaw/agent/runner/messages"
	"time"

	"github.com/JoshPattman/jpf"
)

type encoderInput struct {
	Messages              []messages.Message
	ToolDefs              []agent.ToolDef
	FailedPlugins         []failedPlugin
	Time                  time.Time
	WorkingMemoryLocation string
	WorkingMemory         string
}

type failedPlugin struct {
	Name  string
	Error error
}

// Build the encoder for use with the runner.
func buildEncoder(systemPrompt agent.JsonObject) jpf.Encoder[encoderInput] {
	return &encoder{
		systemPrompt,
	}
}

type encoder struct {
	systemPrompt agent.JsonObject
}

func (e *encoder) BuildInputMessages(input encoderInput) ([]jpf.Message, error) {
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

func (e *encoder) activeState(input encoderInput) agent.JsonObject {
	return agent.JsonObject{
		"description":                  "This is a current state message. You will have been provided with them at previous points in the conversation too, however they have been removed for brevity. This state message is currently up-to-date and active.",
		"active_tools":                 input.ToolDefs,
		"failed_plugins":               input.FailedPlugins,
		"failed_plugins_description":   "This is a list of plugins that failed to load. If a plugin failed to load, all of its tools will be unavailable. If the failed plugins list is empty, happy days!",
		"current_datetime":             input.Time.Format(time.RFC1123),
		"working_memory_file_location": input.WorkingMemoryLocation,
		"working_memory":               input.WorkingMemory,
	}
}

func objectContent(obj agent.JsonObject) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "    ")
	err := enc.Encode(obj)
	if err != nil {
		panic(err)
	}
	return buf.String()
}
