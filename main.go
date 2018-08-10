// Command discordBot
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/sbrow/skirmish"
)

// Trigger is the regexp to match whether to respond to a message or not.
const Trigger = "^!"

// Token holds The API Token for the bot.
var Token string

// ExitStatus is the current exit status.
var ExitStatus = 0

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
	reg := regexp.MustCompile(Trigger)

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Look for the trigger characters
	if reg.MatchString(m.Content) {
		name := reg.ReplaceAllString(m.Content, "")
		card, err := skirmish.Load(name)
		if err != nil && !strings.Contains(err.Error(), "No card found") {
			log.Println(err)
			return
		}
		if card != nil {
			s.ChannelMessageSend(m.ChannelID, card.String())
			return
		}
		rule, err := skirmish.Query(`SELECT rules FROM glossary
WHERE levenshtein(name, $1) <= 3
ORDER BY levenshtein(name, $1) ASC LIMIT 1`, name)
		if err != nil {
			log.Println(err)
			return
		}
		var str *string
		for rule.Next() {
			rule.Scan(&str)
			s.ChannelMessageSend(m.ChannelID, *str)
		}

	}
}
