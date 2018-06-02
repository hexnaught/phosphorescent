package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	phoscom "github.com/jokerdan/phosphorescent/commands"

	"github.com/bwmarrin/discordgo"
)

// == Config =================================================================
type configuration struct {
	BotInfo struct {
		Username string `json:"username"`
		Token    string `json:"token"`
	} `json:"botuser"`
	AVAPIKey string `json:"alphaVantageAPIKey,omitempty"`
}

func loadConfig(file string) (configuration, error) {
	var config configuration
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		return config, err
	}
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	return config, err
}

var config configuration

// == Main ===================================================================

func main() {
	fmt.Println("\nStarting Bot...")
	fmt.Println("+ bwmarrin/discordgo Version:", discordgo.APIVersion)
	// -- Loading Config -------------------------------------------------------
	var err error
	config, err = loadConfig("./config.json")
	if err != nil {
		fmt.Println("! Error with configuration file.", err)
		os.Exit(1)
	}
	fmt.Println("+ Config Loaded!")

	// -- Set Up Bot -----------------------------------------------------------
	fmt.Println("+ Bot Username: ", config.BotInfo.Username)
	disgo, err := discordgo.New("Bot " + config.BotInfo.Token)
	if err != nil {
		fmt.Println("! Error setting up bot.", err)
	}
	disgo.Open()
	defer disgo.Close()
	fmt.Println("+ Connection Now Open!")

	// -- Register Event Handlers ----------------------------------------------
	disgo.AddHandler(commandDispatcher)

	// -- Hold The Program -----------------------------------------------------
	fmt.Printf("Bot is now running.  Press CTRL-C to exit.\n")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

// == Other ==================================================================
// Test of handling messages, will move to message/command parsing within a dir of commands
func commandDispatcher(session *discordgo.Session, message *discordgo.MessageCreate) {

	msg := message

	// If the message is from the bot itself, ignore it
	if msg.Author.ID == session.State.User.ID {
		return
	}

	// Command Prefix
	if msg.Content[0] == ';' {
		msg.Content = msg.Content[1:]
	} else {
		return
	}

	// Safety First
	channel, err := session.State.Channel(message.ChannelID)
	if err != nil {
		return // Could not find channel
	}
	_, err = session.State.Guild(channel.GuildID)
	if err != nil {
		return // Could not find guild
	}

	if msg.Content == "ping" {
		session.ChannelMessageSend(message.ChannelID, "Pong!")
	}
	if msg.Content == "pong" {
		session.ChannelMessageSend(message.ChannelID, "Ping!")
	}

	// Commands/stocks
	if strings.HasPrefix(msg.Content, "stocks") {
		stocks := strings.Fields(strings.TrimPrefix(msg.Content, "stocks"))
		session.ChannelMessageSendEmbed(message.ChannelID, phoscom.GetStock(stocks, config.AVAPIKey))
	}

	if strings.HasPrefix(msg.Content, "trucker") {
		truckerName := strings.Fields(strings.TrimPrefix(msg.Content, "trucker"))
		session.ChannelMessageSendEmbed(message.ChannelID, phoscom.TruckerInfo(truckerName[0]))
	}
}
