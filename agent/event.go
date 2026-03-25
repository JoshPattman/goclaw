package agent

type EventKind string

type Event struct {
	Kind    EventKind
	Content string
}

func E(kind EventKind, content string) Event {
	return Event{kind, content}
}
