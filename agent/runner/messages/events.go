package messages

import (
	"goclaw/agent"

	"github.com/JoshPattman/jpf"
)

func EventsMessage(events ...agent.Event) Message {
	return eventsMessage{events}
}

type eventsMessage struct {
	events []agent.Event
}

func (eventsMessage) Role() jpf.Role {
	return jpf.UserRole
}

func (m eventsMessage) Content() agent.JsonObject {
	eventObjects := make([]agent.JsonObject, len(m.events))
	for i, e := range m.events {
		eventObjects[i] = agent.JsonObject{
			"kind":    e.Kind(),
			"content": e.Content(),
		}
	}
	return agent.JsonObject{
		"explanation": "Some events have occured since your last message. They might be relevant or you might be able to ignore them.",
		"events":      eventObjects,
	}
}
