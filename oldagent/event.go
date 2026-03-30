package oldagent

type EventKind string

type Event interface {
	EventKind() EventKind
	EventData() map[string]any
}

type conversationClippedEvent struct{}

func (conversationClippedEvent) EventKind() EventKind {
	return "conversation_truncated"
}

func (conversationClippedEvent) EventData() map[string]any {
	return map[string]any{
		"explanation": "The conversation has just been truncated (oldest messages were removed). Re-read your system prompt and ensure there is nothing you need to do.",
	}
}

type personalityInstruction struct {
	personality string
}

func (personalityInstruction) EventKind() EventKind {
	return "set_personality"
}

func (e personalityInstruction) EventData() map[string]any {
	return map[string]any{
		"instruction": "Act with this specified personality for further messages in the conversationn the conversation",
		"personality": e.personality,
	}
}

type toolChangeEvent struct {
	tools []Tool
}

func (toolChangeEvent) EventKind() EventKind {
	return "set_available_tools"
}

func (e toolChangeEvent) EventData() map[string]any {
	toolDescs := []map[string]any{
		{
			"tool_name":   doneToolName,
			"description": "Call this tool with no arguments when you think no more actions are required to handle the events. You can only call this tool by itself.",
		},
	}
	for _, t := range e.tools {
		toolDescs = append(toolDescs, map[string]any{
			"tool_name":   t.Name(),
			"description": t.Desc(),
		})
	}
	return map[string]any{
		"explanation":     "The full list of tools has changed, attached is the list of currently avaialable tools",
		"instruction":     "From now on, you may only use the following tools in conversation",
		"available_tools": toolDescs,
	}
}
