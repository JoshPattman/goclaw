package discord

import (
	"fmt"

	"github.com/JoshPattman/cg"

	"github.com/bwmarrin/discordgo"
)

func New(token string) cg.Plugin {
	return &discordPlugin{
		token: token,
	}
}

type discordPlugin struct {
	token string
}

func (p *discordPlugin) Name() string {
	return "discord"
}

func (p *discordPlugin) Load() ([]cg.Tool, <-chan cg.Event, func(), error) {
	sess, err := discordgo.New("Bot " + p.token)
	if err != nil {
		return nil, nil, nil, err
	}
	sess.Identify.Intents =
		discordgo.IntentsGuilds |
			discordgo.IntentsGuildMessages |
			discordgo.IntentsMessageContent
	eventsOut := make(chan cg.Event)
	sess.AddHandler(makeMessageCreateHandler(eventsOut))
	err = sess.Open()
	if err != nil {
		return nil, nil, nil, err
	}
	tools := []cg.Tool{
		&sendMessageTool{sess},
	}
	return tools, eventsOut, func() { sess.Close() }, nil
}

func makeMessageCreateHandler(eventsOut chan<- cg.Event) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		members, err := s.GuildMembers(m.GuildID, "", 1000)
		if err != nil {
			panic(err)
		}
		membersStrs := []discordChannelMember{}
		for _, member := range members {
			membersStrs = append(membersStrs, discordChannelMember{
				member.User.Username,
				member.User.DisplayName(),
			})
		}

		channel, err := s.Channel(m.ChannelID)
		if err != nil {
			panic(err)
		}

		event := discordMessageEvent{
			s.State.User.Username,
			m.ChannelID,
			channel.Name,
			membersStrs,
			m.Author.Username,
			m.Content,
		}

		eventsOut <- event
	}
}

type sendMessageTool struct {
	sess *discordgo.Session
}

func (t *sendMessageTool) Def() cg.ToolDef {
	return cg.ToolDef{
		Name: "discord_send_message",
		Desc: "Send a message to a Discord channel. Args: channel_id (string), content (string)",
	}
}

func (t *sendMessageTool) Call(args map[string]any) (string, error) {
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

type discordMessageEvent struct {
	YourUserName          string
	ChannelID             string
	ChannelName           string
	ChannelMembers        []discordChannelMember
	MessageAuthorUserName string
	MessageContent        string
}

func (e discordMessageEvent) Kind() string {
	return "discord_message_recv"
}

func (e discordMessageEvent) Content() cg.JsonObject {
	return map[string]any{
		"your_user_name":           e.YourUserName,
		"channel_id":               e.ChannelID,
		"channel_name":             e.ChannelName,
		"channel_members":          e.ChannelMembers,
		"message_author_user_name": e.MessageAuthorUserName,
		"message_content":          e.MessageContent,
	}
}

type discordChannelMember struct {
	UserName    string `json:"user_name"`
	DisplayName string `json:"display_name"`
}
