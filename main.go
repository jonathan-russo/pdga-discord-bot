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
	pdgaID, directive, err := parseUserCommand(strings.TrimPrefix(m.Content, TriggerCommand))
	if err != nil {
		log.Printf("Error: %s", err)
		discordErrorReply := fmt.Errorf("Looks like you used the bot wrong! \n %w", err).Error()
		_, err := s.ChannelMessageSend(m.ChannelID, discordErrorReply)
		if err != nil {
			log.Println(err)
		}
		return
	}

	player, err := pdga.NewPlayer(pdgaID)
	if err != nil {
		discordErrorReply := fmt.Errorf("Error retrieving player profile: %w", err).Error()
		_, err = s.ChannelMessageSend(m.ChannelID, discordErrorReply)
		if err != nil {
			log.Println(err)
		}
		return
	}

	if directive == "info" {
		msg := player.Info()
		_, err = s.ChannelMessageSend(m.ChannelID, msg)
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

	if i, err := strconv.Atoi(inputs[0]); err != nil || i < 0 {
		return "", "", errors.New("PDGA number is invalid")
	}

	// Use string here because go slices don't support contains
	allowedDirectives := "info predict_rating"
	if !strings.Contains(allowedDirectives, inputs[1]) {
		return "", "", errors.New("directive invalid")
	}

	return inputs[0], inputs[1], nil
}
