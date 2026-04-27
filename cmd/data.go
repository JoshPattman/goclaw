package main

import (
	"encoding/json"
	"errors"
	"os"
	"path"
)

type Data struct {
	WorkingMemoryLoc string
	AIToken          string
	AIModel          string
	DiscordToken     string
	HTTPMCPs         []HTTPMCPData
	LocalMCPs        []LocalMCPData
	Gmail            bool
	GmailConfigPath  string
	GmailTokenPath   string
	EMLEmail         bool
	EMLEmailPath     string
	EMLEmailUsername string
	EMLEmailAddress  string
	MaxTokens        int
}

type HTTPMCPData struct {
	Address string
	Headers map[string]string
}

type LocalMCPData struct {
	Command []string
}

func UpdateConfig(root string) error {
	err := os.MkdirAll(root, 0775)
	if err != nil {
		return err
	}
	filename := path.Join(root, configFileName)
	data := defaultConfig()
	err = func() error {
		f, err := os.Open(filename)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		} else if err != nil {
			return err
		}
		defer f.Close()
		return json.NewDecoder(f).Decode(&data)
	}()
	if err != nil {
		return err
	}
	return func() error {
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()
		enc := json.NewEncoder(f)
		enc.SetIndent("", "    ")
		return enc.Encode(data)
	}()
}

func LoadData(root string) (Data, error) {
	configPath := path.Join(root, configFileName)
	scratchpadPath := path.Join(root, scratchpadFileName)
	f, err := os.Open(configPath)
	if err != nil {
		return Data{}, err
	}
	defer f.Close()
	config := configDTO{}
	err = json.NewDecoder(f).Decode(&config)
	if err != nil {
		return Data{}, err
	}
	httpMcpDatas := make([]HTTPMCPData, len(config.HTTPMCPDatas))
	for i, m := range config.HTTPMCPDatas {
		httpMcpDatas[i] = HTTPMCPData(m)
	}
	localMcpDatas := make([]LocalMCPData, len(config.LocalMCPDatas))
	for i, m := range config.LocalMCPDatas {
		localMcpDatas[i] = LocalMCPData(m)
	}
	gmailConfigPath := path.Join(root, "gmail_config.json")
	gmailTokenPath := path.Join(root, "gmail_token.json")
	emlIntakeDir := path.Join(root, "eml_intake")
	if config.EMLEmail {
		os.MkdirAll(emlIntakeDir, os.ModePerm)
	}
	return Data{
		scratchpadPath,
		config.AIToken,
		config.AIModel,
		config.DiscordToken,
		httpMcpDatas,
		localMcpDatas,
		config.Gmail,
		gmailConfigPath,
		gmailTokenPath,
		config.EMLEmail,
		emlIntakeDir,
		config.EMLEmailUsername,
		config.EMLEmailAddress,
		config.MaxTokens,
	}, nil
}

const configFileName = "config.json"
const scratchpadFileName = "scratchpad.txt"

type configDTO struct {
	AIToken          string            `json:"ai_token"`
	AIModel          string            `json:"ai_model"`
	DiscordToken     string            `json:"discord_token"`
	HTTPMCPDatas     []httpMcpDataDTO  `json:"http_mcp_servers"`
	LocalMCPDatas    []localMcpDataDTO `json:"local_mcp_servers"`
	Gmail            bool              `json:"gmail"`
	EMLEmail         bool              `json:"eml_email"`
	EMLEmailUsername string            `json:"eml_email_username"`
	EMLEmailAddress  string            `json:"eml_email_address"`
	MaxTokens        int               `json:"max_tokens"`
}

type httpMcpDataDTO struct {
	Address string            `json:"address"`
	Headers map[string]string `json:"headers"`
}

type localMcpDataDTO struct {
	Command []string `json:"command"`
}

func defaultConfig() configDTO {
	return configDTO{
		AIModel:   "gpt-5.2",
		MaxTokens: 16000,
	}
}
