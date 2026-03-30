package agent

type Event interface {
	Kind() string
	Content() JsonObject
}
