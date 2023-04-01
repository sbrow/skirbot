package skirbot

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	// "github.com/sbrow/skirmish"
)

// Prefix is the regexp to match whether to respond to a message or not.
const Prefix = "!"

// query returns results from the database.
func query(m *newMessage) result {
	ret := m.Log()
	content := strings.TrimPrefix(m.Message.Content, Prefix)

	card, err := GetCardByName(content)

	if err != nil {
		ret.Error = err.Error()
		return ret
	}

	ret.Response = card.String()

	return ret
}

// MessageCreate is called every time a new message is created on any channel that the bot has access to.
func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	msg := &newMessage{s, m}

	// Retrieve the name of the channel
	ch, err := msg.Channel()
	if err != nil {
		log.Println(err)
		return
	}

	switch {
	// If the channel is public and the Prefix character was not found, do nothing.
	case ch.Name != "" && !strings.HasPrefix(m.Content, Prefix):
		return
	default:
		response := query(msg)
		go func(r result) {
			data, err := json.Marshal(r)
			if err != nil {
				log.Println(err)
			}
			if r.Guild != "Bot Spam" {
				log.Println(string(data))
			}
		}(response)
		s.ChannelMessageSend(ch.ID, response.Response)
	}
}
