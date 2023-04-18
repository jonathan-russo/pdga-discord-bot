package main

import (
	//   "encoding/json"

	"fmt"
	"log"
	"strings"

	"errors"

	"strconv"

	//   "io/ioutil"
	//   "net/http"
	"os"
	"os/signal"

	"syscall"

	"github.com/bwmarrin/discordgo"

	"github.com/jonathan-russo/pdga-discord-bot/lib/pdga"
)

var (
	Token          string // Discord token for authentication
	TriggerCommand string // Command string used to trigger the bot
)

var UsageGuide = `
Welcome to the PDGA Bot!  This bot can be used to find out information on PDGA players.

Usage: /pdga <command> <pdga-id>

Available Commands:
- info            : This command displays basic information about the player
- predict_rating  : Predict the updated rating for this player at the next ratings update. 
`

func init() {
	TriggerCommand = "/pdga"
	Token = os.Getenv("DISCORD_TOKEN")
}

func main() {

	// Validate Token is present
	if len(Token) == 0 {
		log.Fatal("Discord token not present.  Set env variable 'DISCORD_TOKEN'")
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatalf("Error creating Discord session, %s", err.Error())
	}

	// Register the handleMessage func as a callback for MessageCreate events.
	dg.AddHandler(handleMessage)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		log.Fatalf("error opening connection, %s", err.Error())
	}

	// Wait here until CTRL-C or other term signal is received.
	log.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func handleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself and all messages without the trigger command
	if m.Author.ID == s.State.User.ID || !strings.HasPrefix(m.Content, TriggerCommand) {
		return
	}

	log.Println("Received command: '" + m.Content + "'")

	//Parse user command
	directive, pdgaID, err := parseUserCommand(strings.TrimPrefix(m.Content, TriggerCommand))
	if err != nil {
		log.Printf("Error: %s", err)
		discordErrorReply := fmt.Errorf("***ERROR: Failed to retrieve player profile, %w*** \n %s", err, UsageGuide).Error()
		_, err := s.ChannelMessageSend(m.ChannelID, discordErrorReply)
		if err != nil {
			log.Println(err)
		}
		return
	}

	player, err := pdga.NewPlayer(pdgaID)
	if err != nil {
		discordErrorReply := fmt.Errorf("***ERROR: Failed to retrieve player profile, %w*** \n %s", err, UsageGuide).Error()
		_, err = s.ChannelMessageSend(m.ChannelID, discordErrorReply)
		if err != nil {
			log.Println(err)
		}
		return
	}

	switch directive {
	case "info":
		msg := player.Info()
		_, err = s.ChannelMessageSend(m.ChannelID, msg)
		if err != nil {
			log.Println(err)
		}
	case "predict_rating":
		_, err = s.ChannelMessageSend(m.ChannelID, "Your rating is 7")
		if err != nil {
			log.Println(err)
		}
	default:
		invalidCommandString := fmt.Sprintf("***ERROR: Invalid directive. \n %s", UsageGuide)
		_, err = s.ChannelMessageSend(m.ChannelID, invalidCommandString)
		if err != nil {
			log.Println(err)
		}
	}
}

func parseUserCommand(command string) (string, string, error) {
	inputs := strings.Fields(command)
	if len(inputs) < 2 {
		return "", "", errors.New("invalid number of arguments")
	}

	if i, err := strconv.Atoi(inputs[1]); err != nil || i < 0 {
		return "", "", errors.New("PDGA number is invalid")
	}

	return inputs[0], inputs[1], nil
}
