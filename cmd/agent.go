package main

import (
	"errors"
	"fmt"
	"goclaw/integrations/discord"
	"goclaw/integrations/mail"
	"goclaw/integrations/mcptool"
	"goclaw/services/email"
	"goclaw/services/filemail"
	"goclaw/services/googlemail"
	"log/slog"
	"os"
	"strings"

	"github.com/JoshPattman/cg"
	"github.com/JoshPattman/cg/runner"

	"github.com/JoshPattman/cg/files"

	_ "embed"

	"github.com/JoshPattman/jpf/models"
)

func CreateAgent(data Data) (cg.Agent, error) {
	logger := slog.Default()
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
	fs := files.OSFileSystem()
	fs = files.ListenToWrite(fs, data.WorkingMemoryLoc, func(b []byte) {
		logger.Info("Memory file has been changed")
	})
	ag := runner.New(
		model,
		data.WorkingMemoryLoc,
		fs,
		runner.WithLogger(logger),
		runner.WithMaxTokens(data.MaxTokens),
	)
	ag.AddPlugin(discord.New(data.DiscordToken))
	mcpPlugins := createMCPs(data.HTTPMCPs, data.LocalMCPs)
	for _, p := range mcpPlugins {
		ag.AddPlugin(p)
	}
	if data.Gmail {
		ag.AddPlugin(mail.NewPlugin(
			"gmail",
			func() (email.Client, error) {
				return googlemail.BuildClient(data.GmailConfigPath, data.GmailTokenPath, 8080, func(token string) {
					fmt.Printf("To login to the gmail API, follow this link: %s\n", token)
				})
			},
		))
	}
	if data.EMLEmail {
		ag.AddPlugin(mail.NewPlugin(
			"eml_email",
			func() (email.Client, error) {
				return filemail.NewClient(email.Person{Name: data.EMLEmailUsername, Email: data.EMLEmailAddress}, data.EMLEmailPath)
			},
			mail.WithIngestionFromFolder(data.EMLEmailPath),
		))
	}
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

func createMCPs(httpMCPDatas []HTTPMCPData, localMCPDatas []LocalMCPData) []cg.Plugin {
	plugins := make([]cg.Plugin, 0)
	for _, mcpData := range httpMCPDatas {
		plugins = append(plugins, mcptool.New(mcpData.Address, mcptool.ClientFromHTTP(mcpData.Address, mcpData.Headers)))
	}
	for _, mcpData := range localMCPDatas {
		plugins = append(plugins, mcptool.New(strings.Join(mcpData.Command, " "), mcptool.ClientFromCommand(mcpData.Command)))
	}
	return plugins
}
