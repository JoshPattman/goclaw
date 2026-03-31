package main

import (
	"errors"
	"goclaw/agent"
	"goclaw/agent/files"
	"goclaw/agent/runner"
	"goclaw/integrations/discord"
	"goclaw/integrations/mcptool"
	"log/slog"
	"os"

	_ "embed"

	"github.com/JoshPattman/jpf/models"
)

func CreateAgent(data Data) (agent.Agent, error) {
	err := ensureMemoryFile(data.WorkingMemoryLoc)
	if err != nil {
		return nil, err
	}
	model := models.NewAPIModel(
		models.OpenAI,
		data.AIModel,
		data.AIToken,
		models.WithReasoningEffort(models.LowReasoning),
		models.WithJSONSchema(runner.GetResponseSchema()),
	)
	ag := runner.New(
		model,
		data.WorkingMemoryLoc,
		files.OSFileSystem(),
		runner.WithLogger(slog.Default()),
	)

	discordSession, err := discord.NewDiscordSession(data.DiscordToken, ag.Events())
	if err != nil {
		return nil, err
	}
	err = discordSession.Open()
	if err != nil {
		return nil, err
	}
	ag.AddTools(discordSession.GetTool())
	mcpTools, err := createMCPs(data.HTTPMCPs, data.LocalMCPs)
	if err != nil {
		return nil, err
	}
	ag.AddTools(mcpTools...)
	return ag, nil
}

//go:embed default_memory.txt
var defaultMemoryContent string

func ensureMemoryFile(loc string) error {
	_, err := os.Stat(loc)
	if errors.Is(err, os.ErrNotExist) {
		return os.WriteFile(loc, []byte(defaultMemoryContent), 0644)
	}
	return err
}

func createMCPs(httpMCPDatas []HTTPMCPData, localMCPDatas []LocalMCPData) ([]agent.Tool, error) {
	tools := make([]agent.Tool, 0)
	for _, mcpData := range httpMCPDatas {
		client, err := mcptool.CreateClient(mcpData.Address, mcpData.Headers)
		if err != nil {
			return nil, err
		}
		clientTools, err := mcptool.CreateToolsFromMCP(client)
		if err != nil {
			return nil, err
		}
		tools = append(tools, clientTools...)
	}
	for _, mcpData := range localMCPDatas {
		client, err := mcptool.CreateCommand(mcpData.Command)
		if err != nil {
			return nil, err
		}
		clientTools, err := mcptool.CreateToolsFromMCP(client)
		if err != nil {
			return nil, err
		}
		tools = append(tools, clientTools...)
	}
	return tools, nil
}
