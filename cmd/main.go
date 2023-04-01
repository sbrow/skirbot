// Command skirbot is a Discord bot that queries information about
// cards and rules in the Dreamkeepers: Skirmish card game.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/sbrow/skirbot"
)

// EmptyToken is the default starting Token.
const EmptyToken = ""

// Token holds The API Token for the bot.
var Token string

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
	s.AddHandler(skirbot.MessageCreate)

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
