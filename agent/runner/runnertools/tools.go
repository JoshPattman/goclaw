package runnertools

import (
	"goclaw/agent"
	"goclaw/agent/files"
)

func Plugin(events chan<- agent.Event, fs files.FileSystem) agent.Plugin {
	return runnerPlugin{events: events, fs: fs}
}

type runnerPlugin struct {
	events chan<- agent.Event
	fs     files.FileSystem
}

func (r runnerPlugin) Load() ([]agent.Tool, <-chan agent.Event, func(), error) {
	return []agent.Tool{
		&doneTool{},
		&reminderTool{r.events},
		&listDirectoryTool{r.fs},
		&readFileTool{r.fs, 10000},
		&modifyFileTool{r.fs},
		&deleteFileTool{r.fs},
	}, nil, nil, nil
}

func (r runnerPlugin) Name() string {
	return "internal"
}
