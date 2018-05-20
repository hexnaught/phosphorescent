package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

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

	// parseAVData(config.AVAPIKey)
	// return

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
	if strings.HasPrefix(msg.Content, "stocks") {
		stocks := strings.Fields(strings.TrimPrefix(msg.Content, "stocks"))
		session.ChannelMessageSendEmbed(message.ChannelID, stockEmb(stocks))
	}

}

func sendEmb() *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{},
		// Color:       0x00ff00, // Green
		Color:       0xF6FF93, // Yellow
		Description: "This is a discordgo embed",
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "I am a field",
				Value:  "I am a value",
				Inline: false,
			},
			&discordgo.MessageEmbedField{
				Name:   "I am a second field",
				Value:  "I am a value",
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "I am a Third field",
				Value:  "I am a value",
				Inline: true,
			},
		},
		Image: &discordgo.MessageEmbedImage{
			URL: "https://images.pexels.com/photos/247932/pexels-photo-247932.jpeg?auto=compress&cs=tinysrgb&dpr=2&h=750&w=1260",
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://images.pexels.com/photos/247932/pexels-photo-247932.jpeg?auto=compress&cs=tinysrgb&dpr=2&h=750&w=1260",
		},
		Timestamp: time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
		Title:     "I am an Embed",
	}
	return embed
}

func fetchAVData(xFields []string) AVResp {
	var avData AVResp
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=BATCH_STOCK_QUOTES&symbols=%s&apikey=%s", strings.Join(xFields, ","), config.AVAPIKey)

	// Create Request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return avData
	}

	// Create Client
	client := http.Client{}

	// Run request and get response
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return avData
	}

	// Close when done
	defer resp.Body.Close()

	// Sort out the response
	if err := json.NewDecoder(resp.Body).Decode(&avData); err != nil {
		log.Println(err)
	}

	return avData
}

func stockEmb(symbols []string) *discordgo.MessageEmbed {
	var xFields []*discordgo.MessageEmbedField
	var validSymbols []string
	var xData AVResp

	if len(symbols) > 0 {
		found := map[string]bool{}
		uniqueSymbols := []string{}

		for j := range symbols {
			if found[symbols[j]] == false {
				found[symbols[j]] = true
				uniqueSymbols = append(uniqueSymbols, symbols[j])
			}
		}

		xData = fetchAVData(uniqueSymbols)
	} else {
		xData = fetchAVData([]string{"goog", "aapl", "msft", "nvda"})
	}

	var xPrice float64
	var xTime time.Time
	for i := 0; i < len(xData.MainData); i++ {
		xTime, _ = time.Parse("2006-01-02 15:04:05", xData.MainData[i].Timestamp)
		xPrice, _ = strconv.ParseFloat(xData.MainData[i].Price, 64)
		xFields = append(xFields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("`%s - %s`", xData.MainData[i].Symbol, xTime.Format("15:04:05")),
			Value:  fmt.Sprintf("Price: %.2f\nVolume: %s", xPrice, xData.MainData[i].Volume),
			Inline: true,
		})
		// Only add/show symbols that are valid/found in the API call
		validSymbols = append(validSymbols, xData.MainData[i].Symbol)
	}

	if len(validSymbols) <= 0 {
		xData = fetchAVData([]string{"goog", "aapl", "msft", "nvda"})
	}

	fmt.Println(validSymbols)

	if len(xFields) == 0 {
		return &discordgo.MessageEmbed{
			Color: 0xF6FF93, // Yellow
			Title: "No Stocks Found...",
		}
	}

	embed := &discordgo.MessageEmbed{
		Color:       0xF6FF93, // Yellow
		Description: fmt.Sprintf("Information for %s", strings.Join(validSymbols, ", ")),
		Fields:      xFields,
		Title:       fmt.Sprintf("Stocks - %s", xTime.Format("02-01-2006")),
	}
	return embed
}

//https://www.alphavantage.co/query?function=BATCH_STOCK_QUOTES&symbols=MSFT,FB,AAPL&apikey=demo
//https://www.alphavantage.co/query?function=TIME_SERIES_DAILY&symbol=MSFT&apikey=demo
//https://iextrading.com/developer/docs/#quote
