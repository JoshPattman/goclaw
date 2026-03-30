package main

import (
	"goclaw/agent"
	"goclaw/agent/files"
	"goclaw/agent/runner"
	"goclaw/integrations/discord"
	"log/slog"

	"github.com/JoshPattman/jpf/models"
)

func CreateAgent(data Data) (agent.Agent, error) {
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
	return ag, nil
}
