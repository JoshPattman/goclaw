package runnertools

import (
	"fmt"
	"goclaw/agent"
	"goclaw/agent/files"
)

type modifyFileTool struct {
	fs files.FileSystem
}

func NewModifyFileTool(fs files.FileSystem) agent.Tool {
	return &modifyFileTool{fs: fs}
}

func (t *modifyFileTool) Def() agent.ToolDef {
	return agent.ToolDef{
		Name: "modify_file",
		Desc: "Replace a specific piece of text in a file. Args: 'path' (string), 'old_text' (string), 'new_text' (string). The old_text must appear exactly once. To write a new file, specify old text as an empty string.",
	}
}

func (t *modifyFileTool) Call(args map[string]any) (string, error) {
	pathAny, ok := args["path"]
	if !ok {
		return "", fmt.Errorf("must specify 'path'")
	}
	path, ok := pathAny.(string)
	if !ok {
		return "", fmt.Errorf("'path' must be a string")
	}

	oldTextAny, ok := args["old_text"]
	if !ok {
		return "", fmt.Errorf("must specify 'old_text'")
	}
	oldText, ok := oldTextAny.(string)
	if !ok {
		return "", fmt.Errorf("'old_text' must be a string")
	}

	newTextAny, ok := args["new_text"]
	if !ok {
		return "", fmt.Errorf("must specify 'new_text'")
	}
	newText, ok := newTextAny.(string)
	if !ok {
		return "", fmt.Errorf("'new_text' must be a string")
	}

	err := files.ReplaceText(t.fs, path, oldText, newText)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Successfully updated file: %s", path), nil
}
