package discord

import (
	"fmt"
	"goclaw/agent"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type DiscordSession struct {
	sess   *discordgo.Session
	events chan<- agent.Event
}

func NewDiscordSession(token string, events chan<- agent.Event) (*DiscordSession, error) {
	sess, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	ds := &DiscordSession{
		sess:   sess,
		events: events,
	}

	sess.AddHandler(ds.onMessageCreate)

	return ds, nil
}

func (d *DiscordSession) Open() error {
	return d.sess.Open()
}

func (d *DiscordSession) Close() error {
	return d.sess.Close()
}

func (d *DiscordSession) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	members, err := s.GuildMembers(m.GuildID, "", 1000)
	if err != nil {
		panic(err)
	}
	membersStrs := []string{}
	for _, member := range members {
		if member.User.ID == m.Author.ID {
			membersStrs = append(membersStrs, fmt.Sprintf("%s (this is you)", member.User.DisplayName()))
		} else {
			membersStrs = append(membersStrs, member.User.DisplayName())
		}
	}

	content := fmt.Sprintf(
		"ChannelID: %s (with members %s)\nAuthor: %s\nContent: %s",
		m.ChannelID,
		strings.Join(membersStrs, ", "),
		m.Author.DisplayName(),
		m.Content,
	)

	d.events <- agent.E("discord_message", content)
}

func (d *DiscordSession) GetTool() agent.Tool {
	return &DiscordSendMessageTool{
		sess: d.sess,
	}
}

type DiscordSendMessageTool struct {
	sess *discordgo.Session
}

func (t *DiscordSendMessageTool) Name() string {
	return "discord_send_message"
}

func (t *DiscordSendMessageTool) Desc() string {
	return "Send a message to a Discord channel. Args: channel_id (string), content (string)"
}

func (t *DiscordSendMessageTool) Call(args map[string]any) (string, error) {
	channelIDRaw, ok := args["channel_id"]
	if !ok {
		return "", fmt.Errorf("missing 'channel_id'")
	}

	contentRaw, ok := args["content"]
	if !ok {
		return "", fmt.Errorf("missing 'content'")
	}

	channelID, ok := channelIDRaw.(string)
	if !ok {
		return "", fmt.Errorf("'channel_id' must be a string")
	}

	content, ok := contentRaw.(string)
	if !ok {
		return "", fmt.Errorf("'content' must be a string")
	}

	msg, err := t.sess.ChannelMessageSend(channelID, content)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("message sent (id=%s)", msg.ID), nil
}
