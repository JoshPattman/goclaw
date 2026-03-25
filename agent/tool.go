package agent

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Tool interface {
	Name() string
	Desc() string
	Call(map[string]any) (string, error)
}

func buildChangeToolsPrompt(tools []Tool) string {
	if len(tools) == 0 {
		return "There are currently no tools available."
	}

	var b strings.Builder
	b.WriteString("\n\nThere are currently the following available tools (this message overrides any previous knowledge about tools):\n\n")

	for _, t := range tools {
		b.WriteString(fmt.Sprintf(
			"- %s:\n  %s\n",
			t.Name(),
			t.Desc(),
		))
	}

	b.WriteString(`
When calling tools:
- Use exact tool names
- Provide arguments as arg_name/value pairs
`)

	return b.String()
}

func ParseToolArgs[T any](args map[string]any) (T, error) {
	bs, err := json.Marshal(args)
	if err != nil {
		return *new(T), err
	}
	var t T
	err = json.Unmarshal(bs, &t)
	if err != nil {
		return *new(T), err
	}
	return t, nil
}
