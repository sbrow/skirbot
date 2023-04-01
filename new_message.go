package skirbot

import "github.com/bwmarrin/discordgo"

type result struct {
	Guild    string
	Channel  string
	Author   string
	Message  string
	Response string
	Error    string
}

// newMessage wraps session and Message information for a new message.
type newMessage struct {
	Session *discordgo.Session
	Message *discordgo.MessageCreate
}

// Channel returns the name of the channel, or DM if no name was found.
func (n *newMessage) Channel() (*discordgo.Channel, error) {
	channel, err := n.Session.State.Channel(n.Message.ChannelID)
	if err != nil {
		return n.Session.Channel(n.Message.ChannelID)
	}
	return channel, nil
}

// Log returns the message in a log friendly string format.
func (n *newMessage) Log() result {
	ret := result{Author: n.Message.Author.String()}

	guild, err := n.Guild()
	if err == nil {
		if guild != nil {
			ret.Guild = guild.Name
		}
	}
	ch, err := n.Channel()
	if err == nil {
		if ch.Name != "" {
			ret.Channel = ch.Name
		} else {
			ret.Channel = "Direct Message"
		}
	}
	ret.Message = n.Message.Content
	return ret
}

// Guild returns the Guild that the message was posted in.
func (n *newMessage) Guild() (*discordgo.Guild, error) {
	ch, err := n.Channel()
	if err != nil {
		return nil, err
	}
	guildID := ch.GuildID

	// Attempt to get the guild from the state,
	// If there is an error, fall back to the restapi.
	guild, err := n.Session.State.Guild(guildID)
	if err != nil {
		return n.Session.Guild(guildID)
	}
	return guild, nil
}

// IsDM returns true if a message comes from a DM channel
func (n *newMessage) IsDM() (bool, error) {
	ch, err := n.Channel()
	if err != nil {
		return false, err
	}
	return ch.Name == "DM", nil
}
