package agent

type EventKind string

type Event struct {
	// Kind can be any value, but should sensibly and concisely describe what the type of data is
	Kind EventKind `json:"kind"`
	// Content should be a JSON-encodable data struct
	Content any `json:"content"`
}

func E(kind EventKind, content any) Event {
	return Event{kind, content}
}
