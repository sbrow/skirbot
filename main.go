// Command discordBot
package main

import (
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

// Prefix is the regexp to match whether to respond to a message or not.
const Prefix = "!"

// Token holds The API Token for the bot.
var Token string

// ExitStatus is the current exit status.
// TODO(sbrow): Unused.
var ExitStatus = 0

// Command is a command that a user can execute.
type Command struct {
	f func(*Command, *NewMessage) error
}

// Content returns the Content of the m with the endpoint removed.
func (c *Command) Content(n *NewMessage) string {
	return strings.TrimPrefix(n.Message.Content, Prefix)
}

var Query = &Command{
	f: func(c *Command, n *NewMessage) error {
		content := c.Content(n)

		args := map[string]struct {
			table string
			col   string
		}{
			"card": {"cards", "name"},
			"rule": {"glossary", "rules"},
		}
		query := func(name string) string {
			return fmt.Sprintf("SELECT %s FROM %s Where levenshtein(name, $1) <=2"+
				"ORDER BY levenshtein(name, $1) ASC LIMIT 1", args[name].col, args[name].table)
		}
		var name *string
		err := skirmish.QueryRow(query("card"), content).Scan(&name)
		if name != nil {
			card, err := skirmish.Load(*name)
			if err != nil && !strings.Contains(err.Error(), "No card found") {
				return err
			}
			if card != nil {
				n.Session.ChannelMessageSend(n.Message.ChannelID, card.String())
				return nil
			}
		}
		err = skirmish.QueryRow(query("rule"), content).Scan(&name)
		if err != nil {
			return err
		}
		n.Session.ChannelMessageSend(n.Message.ChannelID, *name)
		return nil
	},
}

func init() {
	flag.StringVar(&Token, "t", "", "The API Token to use for the bot.")
	flag.Parse()
}

func main() {
	defer func() {
		os.Exit(ExitStatus)
	}()

	// End if no token provided
	if Token == "" {
		fmt.Println("No token provided.")
		ExitStatus = 1
		return
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// messageCreate is called every time a new message is created on any channel that the bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	ch, err := s.Channel(m.ChannelID)
	if err != nil {
		log.Println(err)
		return
	}
	name := ch.Name
	if name != "" && !strings.HasPrefix(m.Content, Prefix) {
		return
	}
	if name == "" {
		name = "DM"
	}
	logEntry := fmt.Sprintf("[#%s][%s] \"%s\"", name, m.Author, m.Content)

	msg := &NewMessage{Session: s, Message: m}
	switch {
	default:
		err = Query.f(Query, msg)
	}
	if err != nil {
		// Print error
		defer log.SetPrefix(log.Prefix())
		logEntry += fmt.Sprintf(" ERROR: \"%s\"", err.Error())
	} else {
		// Print message
	}
	log.Println(logEntry)
}

type NewMessage struct {
	Session *discordgo.Session
	Message *discordgo.MessageCreate
}

// isDM returns true if a message comes from a DM channel
func (n *NewMessage) isDM() (bool, error) {
	channel, err := n.Session.State.Channel(n.Message.ChannelID)
	if err != nil {
		if channel, err = n.Session.Channel(n.Message.ChannelID); err != nil {
			return false, err
		}
	}
	return channel.Type == discordgo.ChannelTypeDM, nil
}
