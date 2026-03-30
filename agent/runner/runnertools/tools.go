package runnertools

import (
	"goclaw/agent"
	"goclaw/agent/files"
)

func Tools(events chan<- agent.Event, fs files.FileSystem) []agent.Tool {
	return []agent.Tool{
		&doneTool{},
		&reminderTool{events},
		&listDirectoryTool{fs},
		&readFileTool{fs, 10000},
		&modifyFileTool{fs},
		&deleteFileTool{fs},
	}
}
