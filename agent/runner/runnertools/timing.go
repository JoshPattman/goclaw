package runnertools

import (
	"fmt"
	"goclaw/agent"
	"time"
)

type reminderTool struct {
	events chan<- agent.Event
}

func (t *reminderTool) Def() agent.ToolDef {
	return agent.ToolDef{
		Name: "delayed_event",
		Desc: "Create an event after a delay (like a reminder that will be sent back to you after an amount of time). Args: message (string), delay_seconds (number)",
	}
}

type argData struct {
	Message string  `json:"message"`
	Delay   float64 `json:"delay_seconds"`
}

func (t *reminderTool) Call(args map[string]any) (string, error) {
	ad, err := agent.ParseToolArgs[argData](args)
	if err != nil {
		return "", err
	}
	delay := time.Duration(ad.Delay*1000) * time.Millisecond
	go func() {
		time.Sleep(delay)
		t.events <- reminderEvent{ad.Message}
	}()
	return fmt.Sprintf("event scheduled in %f seconds", ad.Delay), nil
}

type reminderEvent struct {
	message string
}

func (e reminderEvent) Kind() string {
	return "reminder"
}

func (e reminderEvent) Content() agent.JsonObject {
	return map[string]any{
		"reminder_message": e.message,
	}
}
