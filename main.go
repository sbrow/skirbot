// Command discordBot
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/sbrow/skirmish"
)

// EmptyToken is the default starting Token.
const EmptyToken = ""

// Prefix is the regexp to match whether to respond to a message or not.
const Prefix = "!"

// Token holds The API Token for the bot.
var Token string

type Result struct {
	Guild    string
	Channel  string
	Author   string
	Message  string
	Response string
	Error    string
}

// Query returns results from the database.
func Query(m *NewMessage) Result {
	ret := m.Log()
	content := strings.TrimPrefix(m.Message.Content, Prefix)

	var name *string
	err := skirmish.QueryRow("SELECT name FROM cards where levenshtein(name, $1) <=2"+
		"ORDER BY levenshtein(name, $1) ASC LIMIT 1", content).Scan(&name)
	if err != nil {
		ret.Error = err.Error()
		err = skirmish.QueryRow("SELECT long from glossary where levenshtein(name, $1) <= 2", content).Scan(&ret.Response)
		return ret
	} else if name != nil {
		var card skirmish.Card
		card, err = skirmish.Load(*name)
		if card != nil {
			if err != nil && !strings.Contains(err.Error(), "No card found") {
				ret.Error = err.Error()
				return ret
			}
			if card != nil {
				ret.Response = card.String()
				return ret
			}
		}
		err = skirmish.QueryRow("SELECT long from glossary where levenshtein(name, $1) <=2", *name).Scan(&ret.Response)
		if err != nil {
			ret.Error = err.Error()
			return ret
		}
	}
	return ret
}

func init() {
	flag.StringVar(&Token, "t", EmptyToken, "The API Token to use for the bot.")
	flag.Parse()
}

func main() {
	// Terminate if no token was provided
	if Token == EmptyToken {
		fmt.Println("No token provided.")
		os.Exit(1)
	}

	// Create a new Discord session using the provided bot token.
	s, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	s.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	if err = s.Open(); err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	// Cleanly close down the Discord session.
	defer s.Close()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	// Wait for term signal.
	<-sc
}

// messageCreate is called every time a new message is created on any channel that the bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	msg := &NewMessage{s, m}

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
		result := Query(msg)
		go func(r Result) {
			data, err := json.Marshal(r)
			if err != nil {
				log.Println(err)
			}
			if r.Guild != "Bot Spam" {
				log.Println(string(data))
			}
		}(result)
		s.ChannelMessageSend(ch.ID, result.Response)
	}
}

// NewMessage wraps session and Message information for a new message.
type NewMessage struct {
	Session *discordgo.Session
	Message *discordgo.MessageCreate
}

// Channel returns the name of the channel, or DM if no name was found.
func (n *NewMessage) Channel() (*discordgo.Channel, error) {
	channel, err := n.Session.State.Channel(n.Message.ChannelID)
	if err != nil {
		return n.Session.Channel(n.Message.ChannelID)
	}
	return channel, nil
}

// Log returns the message in a log friendly string format.
func (n *NewMessage) Log() Result {
	ret := Result{Author: n.Message.Author.String()}

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
func (n *NewMessage) Guild() (*discordgo.Guild, error) {
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
func (n *NewMessage) IsDM() (bool, error) {
	ch, err := n.Channel()
	if err != nil {
		return false, err
	}
	return ch.Name == "DM", nil
}
