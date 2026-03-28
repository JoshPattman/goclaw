package main

import (
	"goclaw/agent"
	"goclaw/integrations/discord"
	"goclaw/integrations/filesystem"
	"goclaw/integrations/reminders"
	"goclaw/integrations/scratchpad"
	"log/slog"

	"github.com/JoshPattman/jpf/models"
)

func CreateAgent(data Data) (*agent.Agent, error) {
	model := models.NewAPIModel(
		models.OpenAI,
		data.AIModel,
		data.AIToken,
		models.WithReasoningEffort(models.LowReasoning),
	)
	ag := agent.New(model, slog.Default())
	ag.SetPersonality(data.Personality)

	spRead := scratchpad.NewReadScratchPadTool(data.ScratchPad)
	spWrite := scratchpad.NewRewriteScratchPadTool(data.ScratchPad)

	discordSession, err := discord.NewDiscordSession(data.DiscordToken, ag.Events())
	if err != nil {
		return nil, err
	}
	err = discordSession.Open()
	if err != nil {
		return nil, err
	}
	discordSend := discordSession.GetTool()

	reminderTool := reminders.New(ag.Events())

	ag.AddTools(discordSend, reminderTool, spRead, spWrite)
	ag.AddTools(filesystem.Tools()...)
	return ag, nil
}
