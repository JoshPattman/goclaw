package runnertools

import (
	"fmt"
	"goclaw/agent"
	"goclaw/agent/files"
	"strings"
)

type listDirectoryTool struct {
	fs files.FileSystem
}

func (t *listDirectoryTool) Def() agent.ToolDef {
	return agent.ToolDef{
		Name: "list_directory",
		Desc: "List contents of a directory. Arg: 'path' (string). Returns files and subdirectories.",
	}
}

func (t *listDirectoryTool) Call(args map[string]any) (string, error) {
	pathAny, ok := args["path"]
	if !ok {
		return "", fmt.Errorf("must specify a 'path'")
	}

	path, ok := pathAny.(string)
	if !ok {
		return "", fmt.Errorf("'path' must be a string")
	}

	results, err := t.fs.ListDir(path)
	if err != nil {
		return "", err
	}
	return strings.Join(results, "\n"), nil
}
