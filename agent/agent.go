package agent

import (
	_ "embed"
)

// An event is somthing that happens that can trigger the agent to respond.
type Event interface {
	Kind() string
	Content() JsonObject
}

// A tool is somthing the agent can call to perform an action.
type Tool interface {
	Def() ToolDef
	Call(map[string]any) (string, error)
}

// A tool definition specifies how a tool should be used.
type ToolDef struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

// An agent can run (blocking) and respond to events with tool calls.
type Agent interface {
	AddTools(tools ...Tool)
	Events() chan<- Event
	Run() error
}
