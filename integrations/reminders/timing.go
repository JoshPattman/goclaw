package reminders

import (
	"fmt"
	"goclaw/agent"
	"time"
)

// Create a new tool that will send a reminder to the events channel after a number of seconds.
func New(events chan<- agent.Event) agent.Tool {
	return &tool{events}
}

type tool struct {
	events chan<- agent.Event
}

func (t *tool) Name() string {
	return "delayed_event"
}

func (t *tool) Desc() string {
	return "Create an event after a delay (like a reminder that will be sent back to you after an amount of time). Args: message (string), delay_seconds (number)"
}

type argData struct {
	Message string  `json:"message"`
	Delay   float64 `json:"delay_seconds"`
}

func (t *tool) Call(args map[string]any) (string, error) {
	ad, err := agent.ParseToolArgs[argData](args)
	if err != nil {
		return "", err
	}
	delay := time.Duration(ad.Delay*1000) * time.Millisecond
	go func() {
		time.Sleep(delay)
		t.events <- agent.E("delayed_event", ad.Message)
	}()
	return fmt.Sprintf("event scheduled in %f seconds", ad.Delay), nil
}
